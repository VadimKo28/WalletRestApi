package postgres

import (
	"context"
	"os"
	"walet_rest_api/pkg/logging"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context) (*pgxpool.Pool) {
	dsn := os.Getenv("POSTGRES_DATABASE_URL")
	if dsn == "" {
		logging.Logger.Fatal("POSTGRES_DATABASE_URL is not set")
	}

	db, err := pgxpool.New(ctx, dsn)

	if err != nil {
		logging.Logger.WithError(err).Fatal("Unable to create connection pool")
	}

	logging.Logger.Info("Successfully connected to PostgreSQL database")

	return db
}
