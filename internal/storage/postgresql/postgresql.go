package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(host, userName, password, dbName string, port int) (*Storage, error) {
	const op = "storage.postgresql.New"

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable", userName, password, dbName, host, port)

	ctx, cansel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cansel()

	pool, err := pgxpool.New(ctx, connStr)
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
