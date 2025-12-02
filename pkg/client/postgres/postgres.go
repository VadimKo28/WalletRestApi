package postgres

import (
	"context"
	"os"
	"walet_rest_api/pkg/logging"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func NewPool(ctx context.Context) *pgxpool.Pool {
	dsn := os.Getenv("POSTGRES_DATABASE_URL")
	if dsn == "" {
		logging.Logger.Fatal("POSTGRES_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logging.Logger.WithError(err).Fatal("Unable to create connection pool")
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		logging.Logger.WithError(err).Fatal("Unable to ping database")
	}

	logging.Logger.Info("Successfully connected to PostgreSQL database")

	return db
}
