package vfs

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/vmkteam/vfs/db"
)

const AtomTime = "02.01.2006 15:04"

type Folder struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	ParentID *int     `json:"parentId"`
	Folders  []Folder `json:"folders,omitempty"`
}

type FileParams struct {
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
}

type File struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Path        string     `json:"path"`
	PreviewPath string     `json:"previewpath"`
	RelPath     string     `json:"relpath"`
	Size        int        `json:"size"`
	SizeH       []string   `json:"sizeH"`
	Date        string     `json:"date"`
	Type        string     `json:"type"`
	Extension   string     `json:"extension"`
	Params      FileParams `json:"params"`
	Shortpath   string     `json:"shortpath"`
	Width       *int       `json:"width"`
	Height      *int       `json:"height"`
}

func NewFolder(in *db.VfsFolder) *Folder {
	if in == nil {
		return nil
	}

	return &Folder{
		ID:       in.ID,
		Name:     in.Title,
		ParentID: in.ParentFolderID,
	}
}

func NewFile(in *db.VfsFile, webpath, previewpath string) *File {
	if in == nil {
		return nil
	}

	fz := 0
	var hz []string
	if in.FileSize != nil {
		fz = *in.FileSize
		size, unit := getSizeAndUnit(float64(fz))
		hz = []string{fmt.Sprintf("%.*g", 4, size), unit}
	}

	var width, height *int
	fp := FileParams{}
	if in.Params != nil {
		fp.Width = in.Params.Width
		fp.Height = in.Params.Height

		if fp.Width != 0 {
			width = &fp.Width
		}

		if fp.Height != 0 {
			height = &fp.Height
		}
	}

	extension := strings.TrimPrefix(filepath.Ext(in.Path), ".")
	preview := ""
	if isImage(extension) {
		preview = path.Join(previewpath, in.Path)
	}

	return &File{
		ID:          in.ID,
		Name:        in.Title,
		PreviewPath: preview,
		Path:        path.Join(webpath, in.Path),
		RelPath:     strings.TrimSuffix(filepath.Base(in.Path), filepath.Ext(in.Path)),
		Size:        fz,
		SizeH:       hz,
		Date:        in.CreatedAt.Format(AtomTime),
		Type:        in.MimeType,
		Extension:   extension,
		Params:      fp,
		Shortpath:   in.Path,
		Width:       width,
		Height:      height,
	}
}

func NewFullFolder(in *db.VfsFolder, childFolders []db.VfsFolder) *Folder {
	if in == nil {
		return nil
	}

	f := NewFolder(in)
	if len(childFolders) > 0 {
		f.Folders = make([]Folder, len(childFolders))
		for i, cf := range childFolders {
			f.Folders[i] = *NewFolder(&cf)
		}
	}

	return f
}

func getSizeAndUnit(size float64) (newSize float64, unit string) {
	_map := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	base := 1000.0

	i := 0
	unitsLimit := len(_map) - 1
	for size >= base && i < unitsLimit {
		size /= base
		i++
	}
	return size, _map[i]
}

func isImage(ext string) bool {
	for _, imageExt := range []string{"jpg", "jpeg", "gif", "png", "webp", "bmp", "tiff"} {
		if imageExt == ext {
			return true
		}
	}
	return false
}

type HelpUploadItem struct {
	URL    string   `json:"url"`
	Params []string `json:"params"`
}

type HelpUploadResponse struct {
	Temp  HelpUploadItem `json:"temp"`
	Queue HelpUploadItem `json:"queue"`
}

type UrlByHashListResponse struct {
	Hash    string `json:"hash"`
	WebPath string `json:"webPath"`
}
