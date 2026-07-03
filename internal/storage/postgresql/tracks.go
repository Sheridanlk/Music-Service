package postgresql

import (
	"context"
	"fmt"

	"github.com/Sheridanlk/Music-Service/internal/domain/models"
	"github.com/Sheridanlk/Music-Service/internal/storage"
)

func (s *Storage) SaveTrack(ctx context.Context, title, originBucket string) (int64, error) {
	const op = "storage.postgresql.SaveTrack"

	var id int64

	err := s.pool.QueryRow(
		ctx,
		`INSERT INTO tracks (title, origin_bucket) VALUES ($1, $2) RETURNING id`,
		title, originBucket,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: can't insert track: %w", op, err)
	}

	return id, nil
}

func (s *Storage) SetOrginKey(ctx context.Context, id int64, originKey string) error {
	const op = "storage.postgresql.SetOrginKey"

	_, err := s.pool.Exec(
		ctx,
		`UPDATE tracks SET origin_key = $1 WHERE id = $2`,
		originKey, id,
	)
	if err != nil {
		return fmt.Errorf("%s: can't set origin key: %w", op, err)
	}

	return nil
}

func (s *Storage) SetHLS(ctx context.Context, id int64, hlsBucket string, hlsPrefix string) error {
	const op = "storage.postgresql.SetHLS"

	_, err := s.pool.Exec(
		ctx,
		`UPDATE tracks SET hls_bucket = $1, hls_prefix = $2 WHERE id = $3`,
		hlsBucket, hlsPrefix, id,
	)
	if err != nil {
		return fmt.Errorf("%s: can't set hls: %w", op, err)
	}

	return nil
}

func (s *Storage) GetTrack(ctx context.Context, id int64) (models.Track, error) {
	const op = "storage.postgresql.GetTrack"

	var track models.Track

	err := s.pool.QueryRow(
		ctx,
		`SELECT id, title, created_at, origin_bucket, origin_key, hls_bucket FROM tracks WHERE id = $1`,
		id,
	).Scan(&track.ID, &track.Title, &track.CreatedAt, &track.OriginBucket, &track.OriginKey, &track.HLSBucket)
	if err != nil {
		return track, fmt.Errorf("%s: can't get track: %w", op, err)
	}

	return track, nil
}

func (s *Storage) GetHLS(ctx context.Context, id int64) (string, string, error) {
	const op = "storage.postgresql.GetHLS"

	var bucket, prefix *string

	err := s.pool.QueryRow(
		ctx,
		`SELECT hls_bucket, hls_prefix FROM tracks WHERE id = $1`,
		id,
	).Scan(&bucket, &prefix)
	if err != nil {
		return "", "", fmt.Errorf("%s: can't get track hls infornation: %w", op, err)
	}

	return *bucket, *prefix, nil
}

func (s *Storage) GetOriginKey(ctx context.Context, id int64) (string, string, error) {
	const op = "storage.postgresql.GetOriginKey"

	var bucket, key *string

	err := s.pool.QueryRow(
		ctx,
		`SELECT origin_bucket, origin_key FROM tracks WHERE id = $1`,
		id,
	).Scan(&bucket, &key)
	if err != nil {
		return "", "", fmt.Errorf("%s: can't get track origin information: %w", op, err)
	}

	return *bucket, *key, nil
}

func (s *Storage) ListTracks(ctx context.Context, count int, offset int) ([]models.TrackListItem, error) {
	const op = "storage.postgresql.ListTracks"

	tracks := make([]models.TrackListItem, 0, count)

	rows, err := s.pool.Query(
		ctx,
		`SELECT id, title, created_at FROM tracks ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		count, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: can't get tracks: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var track models.TrackListItem
		if err := rows.Scan(&track.ID, &track.Title, &track.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (s *Storage) SetStatusPending(ctx context.Context, id int64) error {
	const op = "storage.postgresql.SetStatusPending"

	res, err := s.pool.Exec(
		ctx,
		`UPDATE tracks SET status = $1 WHERE id = $2 AND status = 'uploading'`,
		storage.StatusPending, id,
	)
	if err != nil {
		return fmt.Errorf("%s: can't set pending status: %w", op, err)
	}
	if rowsAffected := res.RowsAffected(); rowsAffected == 0 {
		return fmt.Errorf("%s: track not found or not in uploading status", op)
	}

	return nil
}

func (s *Storage) SetStatusProcessing(ctx context.Context, id int64) error {
	const op = "storage.postgresql.SetStatusProcessing"

	res, err := s.pool.Exec(
		ctx,
		`UPDATE tracks SET status = $1 WHERE id = $2 AND status = 'pending'`,
		storage.StatusProcessing, id,
	)
	if err != nil {
		return fmt.Errorf("%s: can't set processing status: %w", op, err)
	}
	if rowsAffected := res.RowsAffected(); rowsAffected == 0 {
		return fmt.Errorf("%s: track not found or not in pending status", op)
	}

	return nil
}

func (s *Storage) SetStatusReady(ctx context.Context, id int64) error {
	const op = "storage.postgresql.SetStatusReady"

	res, err := s.pool.Exec(
		ctx,
		`UPDATE tracks SET status = $1 WHERE id = $2 AND status = 'processing'`,
		storage.StatusReady, id,
	)
	if err != nil {
		return fmt.Errorf("%s: can't set ready status: %w", op, err)
	}
	if rowsAffected := res.RowsAffected(); rowsAffected == 0 {
		return fmt.Errorf("%s: track not found or not in processing status", op)
	}

	return nil
}

func (s *Storage) SetStatusError(ctx context.Context, id int64) error {
	const op = "storage.postgresql.SetStatusError"

	_, err := s.pool.Exec(
		ctx,
		`UPDATE tracks SET status = $1 WHERE id = $2`,
		storage.StatusError, id,
	)

	if err != nil {
		return fmt.Errorf("%s: can't set error status: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteTrack(ctx context.Context, id int64) error {
	const op = "storage.postgresql.DeleteTrack"

	_, err := s.pool.Exec(
		ctx,
		`DELETE FROM tracks WHERE id = $1`,
		id,
	)

	if err != nil {
		return fmt.Errorf("%s: can't delete track: %w", op, err)
	}

	return nil
}

func (s *Storage) EditTrack(ctx context.Context, id int64, title string) error {
	const op = "storage.postgresql.EditTrack"

	_, err := s.pool.Exec(
		ctx,
		`UPDATE tracks SET title = $1 WHERE id = $2`,
		title, id,
	)

	if err != nil {
		return fmt.Errorf("%s: can't edit track: %w", op, err)
	}

	return nil
}
