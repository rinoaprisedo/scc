package storage

import "mime/multipart"

// StorageInterface abstracts file storage so a driver (local, s3, ...) can be
// swapped without touching callers. Only LocalStorage is wired up currently —
// main.go constructs it directly.
type StorageInterface interface {
	Upload(file multipart.File, header *multipart.FileHeader, folder string) (string, error)
	Delete(path string) error
	GetURL(path string) string
}

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}
