package vfs

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vmkteam/vfs/db"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-pg/pg/v10"
)

const (
	DefaultHashExtension    = "jpg"
	NamespacePublic         = ""
	defaultModePerm         = os.ModePerm
	defaultHashFileModePerm = 0644
)

var (
	ErrInvalidNamespace = errors.New("invalid namespace")
	ErrInvalidExtension = errors.New("invalid extension")
	ErrInvalidMimeType  = errors.New("invalid mime type")
)

type FileHash struct {
	Hash string
	Ext  string
}

func NewFileHash(hash, ext string) FileHash {
	if ext == "" {
		ext = DefaultHashExtension
	}
	return FileHash{Hash: hash, Ext: ext}
}

func (h FileHash) Dir() string {
	return string(h.Hash[0]) + "/" + h.Hash[1:3]
}

func (h FileHash) File() string {
	return filepath.Join(h.Dir(), h.Hash) + "." + h.Ext
}

type Config struct {
	MaxFileSize      int64
	Path             string
	WebPath          string
	PreviewPath      string
	Database         *pg.Options
	Namespaces       []string
	Extensions       []string
	MimeTypes        []string
	UploadFormName   string
	SaltedFilenames  bool
	SkipFolderVerify bool
}

type VFS struct {
	cfg Config
}

func New(cfg Config) (VFS, error) {
	if !cfg.SkipFolderVerify {
		if _, err := os.Stat(cfg.Path); os.IsNotExist(err) {
			return VFS{}, err
		}
	}

	if cfg.UploadFormName == "" {
		cfg.UploadFormName = "file"
	}

	return VFS{cfg: cfg}, nil
}

func (v VFS) Upload(r io.Reader, relFilename, ns string) error {
	fileDir := filepath.Dir(filepath.Join(v.cfg.Path, ns, relFilename))
	err := os.MkdirAll(fileDir, defaultModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(v.cfg.Path, ns, relFilename))
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return err
	}

	return f.Sync()
}

func (v VFS) Move(ns, currentPath, newPath string) error {
	currentPath = filepath.Join(v.cfg.Path, ns, currentPath)
	newPath = filepath.Join(v.cfg.Path, ns, newPath)

	err := os.MkdirAll(filepath.Dir(newPath), defaultModePerm)
	if err != nil {
		return err
	}

	return os.Rename(currentPath, newPath)
}

func (v VFS) HashUpload(r io.Reader, ns, ext string) (*FileHash, error) {
	if !v.IsValidNamespace(ns) {
		return nil, ErrInvalidNamespace
	}
	if !v.IsValidExtension(ext) {
		return nil, ErrInvalidExtension
	}

	tf, err := os.CreateTemp(v.cfg.Path, "vfs")
	if err != nil {
		return nil, err
	}

	deleteTempFile, tempFilename := true, tf.Name()

	// close and delete file if needed
	defer func() {
		fErr := tf.Close()
		if err == nil && fErr != nil {
			err = fErr
			return
		}

		// delete invalid file
		if deleteTempFile {
			fErr = os.Remove(tempFilename)
			if err == nil && fErr != nil {
				err = fErr
			}
		}
	}()

	// calculate hash
	hash := md5.New()
	wr := io.MultiWriter(hash, tf)
	if _, err := io.Copy(wr, r); err != nil {
		return nil, err
	}

	mType, _ := mimeType(tf)
	if !v.IsValidMimeType(mType) {
		return nil, ErrInvalidMimeType
	}

	hashHex := hex.EncodeToString(hash.Sum(nil)[:16])
	fh := NewFileHash(hashHex, ext)

	// create full path
	err = os.MkdirAll(v.FullDir(ns, fh), defaultModePerm)
	if err != nil {
		return nil, err
	}

	// move temp file to data
	err = os.Rename(tempFilename, v.FullFile(ns, fh))
	if err != nil {
		return nil, err
	}
	err = os.Chmod(v.FullFile(ns, fh), defaultHashFileModePerm)
	if err != nil {
		return nil, err
	}

	// sync file with disk
	if err := tf.Sync(); err != nil {
		return nil, err
	}
	deleteTempFile = false

	return &fh, nil
}

func (v VFS) Path(ns, path string) string {
	return filepath.Join(v.cfg.Path, ns, path)
}

func (v VFS) FullDir(ns string, h FileHash) string {
	return v.Path(ns, h.Dir())
}

func (v VFS) FullFile(ns string, h FileHash) string {
	return v.Path(ns, h.File())
}

func (v VFS) WebHashPath(ns string, h FileHash) string {
	return path.Join(v.cfg.WebPath, ns, h.File())
}

func (v VFS) WebHashPathWithType(ns, fType string, h FileHash) string {
	return path.Join(v.cfg.WebPath, ns, fType, h.File())
}

func (v VFS) WebPath(ns string) string {
	return path.Join(v.cfg.WebPath, ns)
}

func (v VFS) PreviewPath(ns string) string {
	return path.Join(v.cfg.PreviewPath, ns)
}

func (v VFS) IsValidNamespace(ns string) bool {
	if ns == NamespacePublic {
		return true
	}

	for _, n := range v.cfg.Namespaces {
		if n == ns {
			return true
		}
	}

	return false
}

func (v VFS) IsValidExtension(ext string) bool {
	if ext == "" {
		return true
	}

	for _, e := range v.cfg.Extensions {
		if e == ext {
			return true
		}
	}

	return false
}

func (v VFS) IsValidMimeType(mType string) bool {
	if mType == "" {
		return true
	}

	for _, m := range v.cfg.MimeTypes {
		if m == mType {
			return true
		}
	}

	return false
}

type UploadResponse struct {
	Code      int    `json:"-"`                 // http status code
	Error     string `json:"error,omitempty"`   // error message
	Hash      string `json:"hash,omitempty"`    // for hash
	WebPath   string `json:"webPath,omitempty"` // for hash
	FileID    int    `json:"id,omitempty"`      // vfs file id
	Extension string `json:"ext,omitempty"`     // vfs file ext
	Name      string `json:"name,omitempty"`    // vfs file name
	Size      int64  `json:"-"`
}

func (v VFS) writeHashUploadResponse(w http.ResponseWriter, response UploadResponse) error {
	w.WriteHeader(response.Code)
	r, err := json.Marshal(response)
	if err != nil {
		return err
	}

	_, err = w.Write(r)
	return err
}

func (v VFS) uploadFile(r *http.Request, ns, ext, vfsFilename string) UploadResponse {
	var (
		fileSize int64
		rd       io.Reader
		name     string
	)

	// detect PUT or POST usage
	if r.Method == http.MethodPut {
		rd, fileSize = r.Body, r.ContentLength
		if r.Body != nil {
			defer func(b io.ReadCloser) {
				_ = b.Close()
			}(r.Body)
		}
	} else if r.Method == http.MethodPost {
		if err := r.ParseMultipartForm(v.cfg.MaxFileSize); err != nil {
			return UploadResponse{Code: http.StatusInternalServerError, Error: err.Error()}
		}

		file, handler, err := r.FormFile(v.cfg.UploadFormName)
		if err != nil {
			return UploadResponse{Code: http.StatusBadRequest, Error: err.Error()}
		}
		defer func(file multipart.File) {
			_ = file.Close()
		}(file)

		rd, fileSize = file, handler.Size
		ext = strings.TrimPrefix(filepath.Ext(handler.Filename), ".")
		name = strings.TrimSuffix(handler.Filename, filepath.Ext(handler.Filename))
	} else {
		return UploadResponse{Code: http.StatusMethodNotAllowed, Error: "Method not allowed"}
	}

	// validate size
	if fileSize > v.cfg.MaxFileSize {
		return UploadResponse{
			Code:  http.StatusRequestEntityTooLarge,
			Error: fmt.Sprintf("file size exceed %v bytes", v.cfg.MaxFileSize),
		}
	}

	// start normal upload
	if vfsFilename != "" {
		err := v.Upload(rd, vfsFilename, ns)
		if err != nil {
			return UploadResponse{Error: err.Error(), Code: http.StatusBadRequest}
		}

		return UploadResponse{Code: http.StatusOK, Extension: ext, Name: name, Size: fileSize}
	}

	// start hash upload
	fh, err := v.HashUpload(rd, ns, ext)
	if err != nil {
		return UploadResponse{Error: err.Error(), Code: http.StatusBadRequest}
	}

	// write response
	return UploadResponse{Code: http.StatusOK, Hash: fh.Hash, WebPath: v.WebHashPath(ns, *fh), Size: fileSize}
}

func (v VFS) HashUploadHandler(repo *db.VfsRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ns, ext := r.FormValue("ns"), strings.ToLower(r.FormValue("ext"))
		ur := v.uploadFile(r, ns, ext, "")

		if repo != nil && ur.Code == http.StatusOK {
			if ns == "" {
				ns = DefaultNamespace
			}

			if ext == "" {
				ext = DefaultHashExtension
			}

			if err := repo.SaveVfsHash(
				context.Background(),
				&db.VfsHash{Hash: ur.Hash, Namespace: ns, Extension: ext, FileSize: int(ur.Size), CreatedAt: time.Now()},
			); err != nil {
				log.Println("failed to save hash into db err=", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err := v.writeHashUploadResponse(w, ur); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (v VFS) UploadHandler(repo db.VfsRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ns, ext := r.FormValue("ns"), strings.ToLower(r.FormValue("ext"))
		folderId, err := strconv.Atoi(r.FormValue("folderId"))
		if err != nil {
			http.Error(w, "bad folder "+err.Error(), http.StatusBadRequest)
			return
		}

		fl, err := repo.VfsFolderByID(context.Background(), folderId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if fl == nil {
			http.Error(w, "not found", http.StatusInternalServerError)
			return
		}

		// generate temp filename
		tempFile := "temp" + randSeq(16)

		// upload file
		ur := v.uploadFile(r, ns, ext, tempFile)
		if ur.Code == http.StatusOK {
			id, err := v.createFile(repo, fl, ns, tempFile, ur.Name, ur.Extension)
			if err != nil {
				ur.Error = err.Error()
				ur.Code = http.StatusInternalServerError
			} else {
				ur.FileID = id
			}
		}

		if err := v.writeHashUploadResponse(w, ur); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (v VFS) createFile(repo db.VfsRepo, folder *db.VfsFolder, ns, relFilename, name, ext string) (int, error) {
	var (
		params *db.VfsFileParams
		mType  string
		fs     = 0
	)
	if reader, err := os.Open(v.Path(ns, relFilename)); err == nil {
		//check for image
		im, _, err := image.DecodeConfig(reader)
		if err == nil {
			params = &db.VfsFileParams{Height: im.Height, Width: im.Width}
		} else {
			log.Println(err)
		}

		// get file size
		if fi, err := reader.Stat(); err == nil {
			fs = int(fi.Size())
		}

		// detect mime type
		mType, _ = mimeType(reader)

		reader.Close()
	}

	// get last id
	salt := ""
	id, err := repo.NextFileID()
	if err != nil {
		return 0, err
	}

	// set salt
	if v.cfg.SaltedFilenames {
		salt = "_" + randSeq(8)
	}

	// set newFilename
	filename := fmt.Sprintf("%d_%d%s.%s", folder.ID, id, salt, ext) // like 1_9.png
	curYearMonth := time.Now().Format("200601")

	// move temp file to original path
	err = v.Move(ns, relFilename, filepath.Join(curYearMonth, filename))
	if err != nil {
		return 0, err
	}

	// add file to vfs
	vfsFile := db.VfsFile{
		ID:         id,
		FolderID:   folder.ID,
		Title:      name,
		Path:       filepath.Join(curYearMonth, filename),
		Params:     params,
		MimeType:   mType,
		FileSize:   &fs,
		FileExists: true,
		StatusID:   db.StatusEnabled,
		CreatedAt:  time.Now(),
	}

	vf, err := repo.AddVfsFile(context.Background(), &vfsFile)
	if err != nil {
		return 0, err
	}

	return vf.ID, nil
}

func randSeq(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func mimeType(reader *os.File) (string, error) {
	// detect mime type
	_, err := reader.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}
	mt, err := mimetype.DetectReader(reader)
	if err != nil {
		return "", err
	}
	return mt.String(), nil
}
