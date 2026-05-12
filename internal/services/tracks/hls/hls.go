package hls

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Sheridanlk/Music-Service/internal/lib/media"
	"github.com/Sheridanlk/Music-Service/internal/storage"
)

type HlsSegmenter struct {
	log *slog.Logger

	trackProvider TrackProvider
	mediaProvider MediaProvider

	hlsBucket string
}

type TrackProvider interface {
	GetOriginKey(ctx context.Context, id int64) (string, string, error)
	SetHLS(ctx context.Context, id int64, hlsBucket string, hlsPrefix string) error
	SetStatusProcessing(ctx context.Context, id int64) error
	SetStatusReady(ctx context.Context, id int64) error
	SetStatusError(ctx context.Context, id int64) error
}

type MediaProvider interface {
	GetObject(ctx context.Context, bucketName, objectName string, byteRange *storage.ByteRange) (io.ReadCloser, string, int64, error)
	PutObject(ctx context.Context, bucketName, objectName string, r io.Reader, size int64, contentType string) error
}

func New(log *slog.Logger, trackProvider TrackProvider, mediaProvider MediaProvider, hlsBucket string) *HlsSegmenter {
	return &HlsSegmenter{
		log:           log,
		trackProvider: trackProvider,
		mediaProvider: mediaProvider,
		hlsBucket:     hlsBucket,
	}
}

// Hls processes the uploaded track, converts it to HLS format, and uploads HLS files to storage.
func (s *HlsSegmenter) Hls(ctx context.Context, id int64) (err error) {
	const op = "tracks.Hls"

	log := s.log.With(
		slog.String("op", op),
		slog.String("track_id", fmt.Sprintf("%d", id)),
	)

	bucket, originKey, err := s.trackProvider.GetOriginKey(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: failed to get origin key: %w", op, err)
	}

	if err := s.trackProvider.SetStatusProcessing(ctx, id); err != nil {
		return fmt.Errorf("%s: failed to set status processing: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = s.trackProvider.SetStatusError(ctx, id)
		}
	}()

	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("hls-%d-*", id))
	if err != nil {
		return fmt.Errorf("%s: failed to create temp dir: %w", op, err)
	}
	defer os.RemoveAll(tmpDir)

	ext := filepath.Ext(originKey)
	localOriginal := filepath.Join(tmpDir, "original"+ext)
	hlsLocalDir := filepath.Join(tmpDir, "hls")
	os.Mkdir(hlsLocalDir, 0755)

	log.Info("downloading original track", slog.String("bucket", bucket), slog.String("key", originKey))
	body, _, _, err := s.mediaProvider.GetObject(ctx, bucket, originKey, nil)
	if err != nil {
		return fmt.Errorf("%s: failed to download original track: %w", op, err)
	}
	defer body.Close()

	if err := media.WriteToFile(localOriginal, body); err != nil {
		return fmt.Errorf("%s: failed to save original track to local file: %w", op, err)
	}

	log.Info("starting segmentation")

	if err := media.ToHLS(ctx, localOriginal, hlsLocalDir, 4); err != nil {
		return fmt.Errorf("%s: failed to convert to hls: %w", op, err)
	}

	hlsPrefix := media.GenerateTrackHLSKey(id)
	log.Info("uploading generated hls files", slog.String("prefix", hlsPrefix))

	if err := uploadFolder(ctx, s.mediaProvider, s.hlsBucket, hlsLocalDir, hlsPrefix); err != nil {
		return fmt.Errorf("%s: failed to upload hls files: %w", op, err)
	}

	if err := s.trackProvider.SetHLS(ctx, id, s.hlsBucket, hlsPrefix); err != nil {
		return fmt.Errorf("%s: failed to save hls info: %w", op, err)
	}

	if err := s.trackProvider.SetStatusReady(ctx, id); err != nil {
		return fmt.Errorf("%s: failed to set status ready: %w", op, err)
	}

	log.Info("hls processing completed successfully")

	return nil
}

func uploadFolder(ctx context.Context, mediaProvider MediaProvider, hlsBucket, dirPath, prefix string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		localPath := filepath.Join(dirPath, entry.Name())
		objectKey := prefix + entry.Name()

		file, size, err := media.OpenFile(localPath)
		if err != nil {
			return err
		}

		ct := media.DetectContentType(entry.Name())

		err = mediaProvider.PutObject(ctx, hlsBucket, objectKey, file, size, ct)
		file.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
