package postgresql

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Sheridanlk/Music-Service/internal/domain/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(host, userName, password, dbName string, port int) (*Storage, error) {
	const op = "storage.postgresql.New"

	connString := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(userName, password),
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   dbName,
	}

	cfg, err := pgxpool.ParseConfig(connString.String())
	if err != nil {
		return nil, fmt.Errorf("%s: can't parse connection string: %w", op, err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	ctx, cansel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cansel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: can't connect to database:%w", op, err)
	}

	ctxPing, canselPing := context.WithTimeout(context.Background(), 2*time.Second)
	defer canselPing()
	if err := pool.Ping(ctxPing); err != nil {
		return nil, fmt.Errorf("%s: can't ping database:%w", op, err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}

func (s *Storage) Create(ctx context.Context, title, originBucket, originKey string) (int64, error) {
	const op = "storage.postgresql.Create"

	var id int64

	err := s.pool.QueryRow(
		ctx,
		`INSERT INTO tracks (title, origin_bucket, origin_key) VALUES ($1, $2, $3) RETURNING id`,
		title, originBucket, originKey,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: can't insert track: %w", op, err)
	}

	return id, nil
}

func (s *Storage) Get(ctx context.Context, id int64) (models.Track, error) {
	const op = "storage.postgresql.Get"

	var track models.Track

	err := s.pool.QueryRow(
		ctx,
		`SELECT id, title, created_at, origin_bucket, origin_key, hls_bucket FROM tracks WHERE id = $1`,
		id,
	).Scan(&track.ID, &track.Title, &track.CreatedAt, &track.OriginBucvket, &track.OriginKey, &track.HLSBucket)
	if err != nil {
		return track, fmt.Errorf("%s: can't get track: %w", op, err)
	}

	return track, nil
}
