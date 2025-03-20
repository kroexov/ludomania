package vfs

import (
	"context"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"

	"github.com/vmkteam/vfs/db"

	"github.com/go-pg/pg/v10"
	"github.com/vmkteam/zenrpc/v2"
)

var (
	ErrInternal     = httpAsRpcError(http.StatusInternalServerError)
	ErrNotFound     = httpAsRpcError(http.StatusNotFound)
	ErrInvalidSort  = zenrpc.NewStringError(http.StatusBadRequest, "invalid sort field")
	ErrInvalidInput = zenrpc.NewStringError(http.StatusBadRequest, "invalid user input")
)

var filenameRegex = regexp.MustCompile(`^([0-9a-z_-])+\.([0-9a-z])+$`)

func httpAsRpcError(code int) *zenrpc.Error {
	return zenrpc.NewStringError(code, http.StatusText(code))
}

func InternalError(err error) *zenrpc.Error {
	return zenrpc.NewError(http.StatusInternalServerError, err)
}

//go:generate zenrpc
type Service struct {
	zenrpc.Service
	dbc  *pg.DB
	repo db.VfsRepo
	vfs  VFS
}

func NewService(repo db.VfsRepo, vfs VFS, dbc *pg.DB) Service {
	return Service{repo: repo, vfs: vfs, dbc: dbc}
}

func (s Service) folderByID(ctx context.Context, id int) (*db.VfsFolder, error) {
	dbc, err := s.repo.VfsFolderByID(ctx, id, s.repo.FullVfsFolder())
	if err != nil {
		return nil, InternalError(err)
	} else if dbc == nil {
		return nil, ErrNotFound
	}
	return dbc, nil
}

// GetFolder returns Folder with sub folders.
//
//zenrpc:rootFolderId=1
//zenrpc:404 Folder not found
func (s Service) GetFolder(ctx context.Context, rootFolderId int) (*Folder, error) {
	dbf, err := s.folderByID(ctx, rootFolderId)
	if err != nil {
		return nil, err
	}

	childFolders, err := s.repo.VfsFoldersByFilters(ctx, &db.VfsFolderSearch{ParentFolderID: &dbf.ID}, db.PagerNoLimit)
	if err != nil {
		return nil, InternalError(err)
	}

	return NewFullFolder(dbf, childFolders), nil
}

// GetFolderBranch returns Folder branch.
func (s Service) GetFolderBranch(ctx context.Context, folderId int) ([]Folder, error) {
	dbf, err := s.folderByID(ctx, folderId)
	if err != nil {
		return nil, err
	}

	list, err := s.repo.FolderBranch(ctx, dbf.ID)
	if err != nil {
		return nil, InternalError(err)
	}

	folders := make([]Folder, 0, len(list))
	for i := 0; i < len(list); i++ {
		folders = append(folders, *NewFolder(&list[i]))
	}
	return folders, nil
}

// GetFiles returns list of files.
//
//zenrpc:folderId root folder id
//zenrpc:query file name
//zenrpc:sortField="createdAt" createdAt, title or fileSize
//zenrpc:isDescending=true asc = false, desc = true
//zenrpc:page=0 current page
//zenrpc:pageSize=100 current pageSize
func (s Service) GetFiles(ctx context.Context, folderId int, query *string, sortField string, isDescending bool, page, pageSize int) ([]File, error) {
	dbf, err := s.folderByID(ctx, folderId)
	if err != nil {
		return nil, err
	}

	if sortField != "createdAt" && sortField != "title" && sortField != "fileSize" {
		return nil, ErrInvalidSort
	}

	// set sort
	sort := db.SortField{Column: sortField, Direction: db.SortAsc}
	if isDescending {
		sort.Direction = db.SortDesc
	}

	search := (&db.VfsFileSearch{FolderID: &dbf.ID}).WithQuery(query)
	list, err := s.repo.VfsFilesByFilters(ctx, search, db.Pager{Page: page, PageSize: pageSize}, db.WithSort(sort))
	if err != nil {
		return nil, InternalError(err)
	}

	files := make([]File, 0, len(list))
	for i := 0; i < len(list); i++ {
		// nolint
		files = append(files, *NewFile(&list[i], s.vfs.WebPath(""), s.vfs.PreviewPath(""))) // TODO ns?
	}
	return files, nil
}

// CountFiles returns count of files.
//
//zenrpc:folderId root folder id
//zenrpc:query file name
func (s Service) CountFiles(ctx context.Context, folderId int, query *string) (int, error) {
	search := (&db.VfsFileSearch{FolderID: &folderId}).WithQuery(query)
	count, err := s.repo.CountVfsFiles(ctx, search)
	if err != nil {
		return 0, InternalError(err)
	}

	return count, nil
}

// MoveFiles move files to destination folder.
//
//zenrpc:400 empty file ids
func (s Service) MoveFiles(ctx context.Context, fileIds []int64, destinationFolderId int) (bool, error) {
	fl, err := s.folderByID(ctx, destinationFolderId)
	if err != nil {
		return false, err
	}

	if len(fileIds) == 0 {
		return false, ErrInvalidInput
	}

	r, err := s.repo.UpdateFilesFolder(ctx, fileIds, fl.ID)
	if err != nil {
		return false, InternalError(err)
	}

	return r, nil
}

// DeleteFiles remove files.
func (s Service) DeleteFiles(ctx context.Context, fileIds []int64) (bool, error) {
	if len(fileIds) == 0 {
		return false, ErrInvalidInput
	}

	r, err := s.repo.DeleteVfsFiles(ctx, fileIds)
	if err != nil {
		return false, InternalError(err)
	}

	return r, nil
}

// SetFilePhysicalName renames File on server.
func (s Service) SetFilePhysicalName(ctx context.Context, fileId int, name string) (bool, error) {
	if fileId == 0 || name == "" {
		return false, ErrInvalidInput
	}

	f, err := s.repo.VfsFileByID(ctx, fileId, db.EnabledOnly())
	if err != nil {
		return false, InternalError(err)
	} else if f == nil {
		return false, ErrNotFound
	}

	if !filenameRegex.MatchString(name) {
		return false, httpAsRpcError(http.StatusNotAcceptable)
	}

	oldPath, newPath := f.Path, filepath.Join(filepath.Dir(f.Path), name)
	if _, err := os.Stat(s.vfs.Path(NamespacePublic, newPath)); err == nil {
		return false, httpAsRpcError(http.StatusConflict)
	}

	// update path and move file in transaction
	err = s.dbc.RunInTransaction(ctx, func(tx *pg.Tx) error {
		txr := s.repo.WithTransaction(tx)

		f.Path = newPath
		_, err := txr.UpdateVfsFile(ctx, f, db.WithColumns(db.Columns.VfsFile.Path))
		if err != nil {
			return err
		}

		return s.vfs.Move(NamespacePublic, oldPath, newPath)
	})

	if err != nil {
		return false, InternalError(err)
	}

	return true, nil
}

// SearchFolderByFileId return Folder by File id.
func (s Service) SearchFolderByFileId(ctx context.Context, fileId int) (*Folder, error) {
	if fileId == 0 {
		return nil, ErrInvalidInput
	}

	f, err := s.repo.VfsFileByID(ctx, fileId, s.repo.FullVfsFile())
	if err != nil {
		return nil, InternalError(err)
	} else if f == nil {
		return nil, ErrNotFound
	}

	return NewFolder(f.Folder), nil
}

// SearchFolderByFile return Folder by File name.
func (s Service) SearchFolderByFile(ctx context.Context, filename string) (*Folder, error) {
	if filename == "" {
		return nil, ErrInvalidInput
	}

	list, fn := path.Split(filename)
	searchPath := path.Join(path.Base(list), fn)

	f, err := s.repo.OneVfsFile(ctx, &db.VfsFileSearch{Path: &searchPath}, s.repo.FullVfsFile())
	if err != nil {
		return nil, InternalError(err)
	} else if f == nil {
		return nil, ErrNotFound
	}

	return NewFolder(f.Folder), nil
}

// GetFavorites return favorites list.
func (s Service) GetFavorites(ctx context.Context) ([]Folder, error) {
	b := true
	list, err := s.repo.VfsFoldersByFilters(ctx, &db.VfsFolderSearch{IsFavorite: &b}, db.PagerNoLimit)
	if err != nil {
		return nil, InternalError(err)
	}

	folders := make([]Folder, 0, len(list))
	for i := 0; i < len(list); i++ {
		folders = append(folders, *NewFolder(&list[i]))
	}
	return folders, nil
}

// ManageFavorites manage favorite virtual folders.
func (s Service) ManageFavorites(ctx context.Context, folderId int, isInFavorites bool) (bool, error) {
	if folderId == 0 || folderId == 1 {
		return false, ErrInvalidInput
	}

	f, err := s.folderByID(ctx, folderId)
	if err != nil {
		return false, InternalError(err)
	} else if f == nil {
		return false, ErrNotFound
	}

	f.IsFavorite = &isInFavorites
	return s.repo.UpdateVfsFolder(ctx, f, db.WithColumns(db.Columns.VfsFolder.IsFavorite))
}

// CreateFolder create virtual folder.
func (s Service) CreateFolder(ctx context.Context, rootFolderId int, name string) (bool, error) {
	f, err := s.folderByID(ctx, rootFolderId)
	if err != nil {
		return false, err
	}

	if name == "" {
		return false, ErrInvalidInput
	}

	dbf := &db.VfsFolder{
		ParentFolderID: &f.ID,
		Title:          name,
		CreatedAt:      time.Now(),
		StatusID:       db.StatusEnabled,
	}

	if _, err = s.repo.AddVfsFolder(ctx, dbf); err != nil {
		return false, InternalError(err)
	}

	return false, nil
}

// DeleteFolder removes Folder.
func (s Service) DeleteFolder(ctx context.Context, folderId int) (bool, error) {
	f, err := s.folderByID(ctx, folderId)
	if err != nil {
		return false, err
	}

	if folderId == 1 { // root
		return false, ErrInvalidInput
	}

	if _, err = s.repo.DeleteVfsFolder(ctx, f.ID); err != nil {
		return false, InternalError(err)
	}

	return true, nil
}

// MoveFolder move Folder to destination folder.
func (s Service) MoveFolder(ctx context.Context, folderId, destinationFolderId int) (bool, error) {
	if folderId == 1 || folderId == 0 || destinationFolderId == 0 || folderId == destinationFolderId {
		return false, ErrInvalidInput
	}

	// validate
	fl, err := s.folderByID(ctx, folderId)
	if err != nil {
		return false, InternalError(err)
	}

	dfl, err := s.folderByID(ctx, destinationFolderId)
	if err != nil {
		return false, InternalError(err)
	}

	if fl == nil || dfl == nil {
		return false, ErrNotFound
	}

	// check recursive path
	if pathList, err := s.repo.FolderBranch(ctx, destinationFolderId); err != nil {
		return false, InternalError(err)
	} else {
		for i := len(pathList) - 1; i >= 0; i-- {
			p := pathList[i]
			if p.ID == folderId {
				return false, httpAsRpcError(http.StatusConflict)
			}
		}
	}

	// move
	fl.ParentFolderID = &dfl.ID
	r, err := s.repo.UpdateVfsFolder(ctx, fl, db.WithColumns(db.Columns.VfsFolder.ParentFolderID))
	if err != nil {
		return false, InternalError(err)
	}

	return r, nil
}

// RenameFolder change Folder name.
func (s Service) RenameFolder(ctx context.Context, folderId int, name string) (bool, error) {
	if folderId == 0 || folderId == 1 || name == "" {
		return false, ErrInvalidInput
	}

	f, err := s.folderByID(ctx, folderId)
	if err != nil {
		return false, InternalError(err)
	} else if f == nil {
		return false, ErrNotFound
	}

	f.Title = name
	return s.repo.UpdateVfsFolder(ctx, f, db.WithColumns(db.Columns.VfsFolder.Title))
}

// HelpUpload returns a uploader help info.
func (s Service) HelpUpload() HelpUploadResponse {
	return HelpUploadResponse{
		Temp: HelpUploadItem{
			URL:    "/vfs/upload/hash",
			Params: []string{s.vfs.cfg.UploadFormName},
		},
		Queue: HelpUploadItem{
			URL:    "/vfs/upload/file",
			Params: []string{s.vfs.cfg.UploadFormName, "folderId"},
		},
	}
}

// UrlByHash get Url by hash, namespace and media type
//
//zenrpc:hash media hash
//zenrpc:namespace media namespace
//zenrpc:mediaType type of media (possible values: small, medium, big, empty string)
func (s Service) UrlByHash(_ context.Context, hash, namespace, mediaType string) (string, error) {
	return s.vfs.WebHashPathWithType(namespace, mediaType, NewFileHash(hash, "")), nil
}

// UrlByHashList get Urls by hash list, with namespace and media type
//
//zenrpc:hashList media hash list
//zenrpc:namespace media namespace
//zenrpc:mediaType type of media (possible values: small, medium, big, empty string)
func (s Service) UrlByHashList(ctx context.Context, hashList []string, namespace, mediaType string) ([]UrlByHashListResponse, error) {
	var resp []UrlByHashListResponse
	for _, hash := range hashList {
		url, _ := s.UrlByHash(ctx, hash, namespace, mediaType)
		resp = append(resp, UrlByHashListResponse{Hash: hash, WebPath: url})
	}

	return resp, nil
}

// DeleteHash delete file by namespace and hash.
//
//zenrpc:namespace media namespace
//zenrpc:hash media hash
//zenrpc:ext media extension
//zenrpc:404 File not found by hash
func (s Service) DeleteHash(ctx context.Context, namespace, hash string) (bool, error) {
	vfsHash, err := s.repo.VfsHashByID(ctx, hash, namespace)
	if err != nil {
		return false, InternalError(err)
	}
	if vfsHash == nil {
		return false, ErrNotFound
	}

	_, err = s.repo.DeleteVfsHash(ctx, hash, namespace)
	if err != nil {
		return false, InternalError(err)
	}

	fileName := s.vfs.FullFile(namespace, NewFileHash(vfsHash.Hash, vfsHash.Extension))
	_, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		return false, ErrNotFound
	} else if err != nil {
		return false, InternalError(err)
	}
	err = os.Remove(fileName)
	if err != nil {
		return false, InternalError(err)
	}

	return true, nil
}
