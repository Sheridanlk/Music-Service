package media

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
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	log         *slog.Logger
	minioclient *minio.Client
}

func New(log *slog.Logger, endpoint string, accessKeyID string, secretAccessKey string, useSSL bool) (*MinioStorage, error) {
	const op = "storage.minio.New"
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_ = mime.AddExtensionType(".m3u8", "application/vnd.apple.mpegurl")
	_ = mime.AddExtensionType(".aac", "audio/aac")
	return &MinioStorage{
		minioclient: client,
		log:         log,
	}, nil
}

func (s *MinioStorage) PutObject(ctx context.Context, bucketName, objectName string, r io.Reader, size int64, contentType string) error {
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

func (s *MinioStorage) GetObject(ctx context.Context, bucketName, objectName string, byteRange *storage.ByteRange) (io.ReadCloser, string, int64, error) {
	const op = "storage.minio.Download"

	opts := minio.GetObjectOptions{}
	if byteRange != nil {
		opts.SetRange(byteRange.Start, byteRange.End)
	}

	obj, err := s.minioclient.GetObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return nil, "", 0, fmt.Errorf("%s: can't download object: %w", op, err)
	}

	st, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, "", 0, fmt.Errorf("%s: can't get object stats: %w", op, err)
	}

	ct := contentTypeByExt(objectName)

	return obj, ct, st.Size, nil
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
