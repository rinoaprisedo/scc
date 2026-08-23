package storage

import (
	"io"
	"mime/multipart"
)

// StorageInterface abstracts file storage so a driver (local, s3, ...) can be
// swapped without touching callers. main.go constructs S3Storage directly —
// LocalStorage still exists and satisfies the interface, but every upload
// path (peserta, users, settings, qris_cross_border) needs files to survive
// a redeploy, which a local disk path doesn't (see backend/Dockerfile's note
// on uploads not being a mounted volume), so it's not currently used.
type StorageInterface interface {
	Upload(file multipart.File, header *multipart.FileHeader, folder string) (string, error)
	Delete(path string) error
	GetURL(path string) string
	// Open returns the stored file's contents for server-side reads (e.g.
	// bundling several files into a ZIP) — unlike GetURL, which only builds
	// a link for the browser to fetch directly.
	Open(path string) (io.ReadCloser, error)
}

var allowedMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"application/pdf": true,
}
