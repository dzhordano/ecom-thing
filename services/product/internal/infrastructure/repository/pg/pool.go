package pg

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MustNewPGXPool(ctx context.Context, dsn string) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	return pool
}
