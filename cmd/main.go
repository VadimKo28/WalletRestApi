package main

import (
	"context"
	"walet_rest_api/pkg/client/postgres"
	"walet_rest_api/pkg/logging"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		logging.Logger.WithError(err).Warn("Failed to load .env file")
	}

	db := postgres.NewPool(context.TODO())

	defer db.Close()
}