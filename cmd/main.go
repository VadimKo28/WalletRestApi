package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"walet_rest_api/internal/config"
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

	godotenv.Load()

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	db := postgres.NewPool(ctx)

	defer db.Close()

	storage := walletdb.NewWalletDB(db, logger)

	service := wallet.NewService(storage)

	h := handler.NewHandlers(service, logger)

	router := gin.Default()
	h.RegisterRoutes(router)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("failed to run HTTP server")
		}
	}()

	logger.Infof("HTTP server started on %s", cfg.HTTPAddr)

	<-ctx.Done()
	logger.Info("Shutting down HTTP server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("HTTP server forced to shutdown")
	} else {
		logger.Info("HTTP server stopped gracefully")
	}
}
