//nolint:unused
package vt

import (
	"path"

	"apisrv/pkg/db"
)

const (
	mediaNormalImage = "normal"
	mediaBigImage    = "big"
	mediaMediumImage = "medium"
	mediaSmallImage  = "small"
	media128Image    = "128"
	media256Image    = "256"
	media512Image    = "512"
	media768Image    = "768"
	media1024Image   = "1024"
	media2048Image   = "2048"
)

type VfsFileSummary struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type VfsHashImage struct {
	Hash    string `json:"hash"`
	WebPath string `json:"webPath"`
}

var WebPath string

// NewVfsFileSummary converts db.VfsFile to VfsFileSummary.
func NewVfsFileSummary(in *db.VfsFile) *VfsFileSummary {
	if in == nil {
		return nil
	}

	return &VfsFileSummary{
		ID:   in.ID,
		Name: in.Title,
		Path: WebPath + in.Path,
	}
}

// newVfsHashImage converts string to VfsHashImage.
func newVfsHashImage(in string) *VfsHashImage {
	if in == "" {
		return nil
	}

	return &VfsHashImage{
		Hash:    in,
		WebPath: mediaImage(in, media256Image),
	}
}

// newVfsHashImages converts []string to []VfsHashImage.
func newVfsHashImages(in []string) (out []VfsHashImage) {
	out = make([]VfsHashImage, len(in))

	for i, value := range in {
		out[i] = *newVfsHashImage(value)
	}

	return
}

// mediaImage returns full path for vfs image.
func mediaImage(hash, size string) string {
	if len(hash) != 32 {
		return ""
	}

	return WebPath + path.Join(
		size,
		hash[:1],
		hash[1:3],
		hash+".jpg",
	)
}
