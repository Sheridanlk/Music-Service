package postgresql

import (
	"context"
	"fmt"
	"net/url"
	"time"

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
