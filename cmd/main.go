package main

import (
	"context"

	"walet_rest_api/internal/domain/wallet"
	walletdb "walet_rest_api/internal/domain/wallet/db"
	"walet_rest_api/internal/handler"
	"walet_rest_api/pkg/client/postgres"
	"walet_rest_api/pkg/logging"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	logger := logging.GetLogger()

	if err := godotenv.Load(); err != nil {
		logger.WithError(err).Warn("Failed to load .env file")
	}

	ctx := context.Background()
	db := postgres.NewPool(ctx)

	if err := db.Ping(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to ping database")
	}

	defer db.Close()

	storage := walletdb.NewWalletDB(db, logger)

	service := wallet.NewService(storage)

	h := handler.NewHandlers(service, logger)

	router := gin.Default()
	h.RegisterRoutes(router)

	if err := router.Run(":3010"); err != nil {
		logger.WithError(err).Fatal("failed to run HTTP server")
	}
}
