package minio

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"path/filepath"
	"strings"

	"github.com/Sheridanlk/Music-Service/internal/storage"
	"github.com/minio/minio-go/v7"
)

type MinioStorage struct {
	minioclient *minio.Client
	log         *slog.Logger
}

func New(log *slog.Logger, client *minio.Client) *MinioStorage {
	_ = mime.AddExtensionType(".m3u8", "application/vnd.apple.mpegurl")
	_ = mime.AddExtensionType(".aac", "audio/aac")
	return &MinioStorage{
		minioclient: client,
		log:         log,
	}
}

func (s *MinioStorage) Upload(ctx context.Context, bucketName, objectName string, r io.Reader, size int64, contentType string) error {
	const op = "storage.minio.Upload"

	_, err := s.minioclient.PutObject(
		ctx,
		bucketName,
		objectName,
		r,
		size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return fmt.Errorf("%s: can't upload object to minio: %w", op, err)
	}

	return nil
}

func (s *MinioStorage) Download(ctx context.Context, bucketName, objectName string, byteRange *storage.ByteRannge) (io.ReadCloser, string, error) {
	const op = "storage.minio.Download"

	opts := minio.GetObjectOptions{}
	if byteRange != nil {
		opts.SetRange(byteRange.Start, byteRange.End)
	}

	obj, err := s.minioclient.GetObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return nil, "", fmt.Errorf("%s: can't download object: %w", op, err)
	}

	ct := contentTypeByExt(objectName)

	return obj, ct, nil
}

func contentTypeByExt(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		return "application/octet-stream"
	}
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}
	switch ext {
	case ".m3u8":
		return "application/vnd.apple.mpegurl"
	case ".aac":
		return "audio/aac"
	default:
		return "application/octet-stream"
	}
}
