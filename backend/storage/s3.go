package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

// S3Storage stores files in an S3-compatible object store (UpCloud Object
// Storage) instead of local disk — every photo upload across the app
// (peserta KTP scans, qris_cross_border proof images, user avatars, settings
// branding/content) goes through whichever StorageInterface impl main.go
// constructs, so this is the single place that switches all of them over.
type S3Storage struct {
	Client  *s3.Client
	Bucket  string
	BaseURL string // public URL prefix objects are served from: "<endpoint>/<bucket>"
	MaxSize int64
}

func NewS3Storage(endpoint, region, accessKey, secretKey, bucket string, maxSize int64) *S3Storage {
	client := s3.New(s3.Options{
		Region:       region,
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		BaseEndpoint: aws.String(endpoint),
		// Most S3-compatible providers (UpCloud included) don't support
		// AWS's virtual-hosted-style bucket subdomains for arbitrary
		// endpoints — path-style (`<endpoint>/<bucket>/<key>`) is what
		// actually works against a custom BaseEndpoint.
		UsePathStyle: true,
	})

	return &S3Storage{
		Client:  client,
		Bucket:  bucket,
		BaseURL: strings.TrimRight(endpoint, "/") + "/" + bucket,
		MaxSize: maxSize,
	}
}

func (s *S3Storage) Upload(file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	if header.Size > s.MaxSize {
		return "", fmt.Errorf("file exceeds max size of %d bytes", s.MaxSize)
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	contentType := http.DetectContentType(buf[:n])
	if !allowedMimeTypes[contentType] {
		return "", fmt.Errorf("unsupported file type: %s", contentType)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	ext := filepath.Ext(header.Filename)
	key := filepath.ToSlash(filepath.Join(folder, uuid.New().String()+ext))

	_, err = s.Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		// Every upload in this app is served back as a plain, unauthenticated
		// URL (same assumption LocalStorage's /uploads static route made) —
		// there's no signed-URL/proxy layer here, so the object must be
		// publicly readable or GetURL's link would 403 in the browser.
		ACL: types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", err
	}

	return key, nil
}

func (s *S3Storage) Open(path string) (io.ReadCloser, error) {
	out, err := s.Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *S3Storage) Delete(path string) error {
	_, err := s.Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(path),
	})
	return err
}

func (s *S3Storage) GetURL(path string) string {
	return s.BaseURL + "/" + strings.TrimLeft(path, "/")
}
