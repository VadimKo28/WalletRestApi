package postgres

import (
	"context"
	"os"
	"testing"

	"walet_rest_api/pkg/logging"
)

func TestNewPool_MissingEnv_TriggersFatal(t *testing.T) {
	origDSN, hadDSN := os.LookupEnv("POSTGRES_DATABASE_URL")
	if hadDSN {
		defer os.Setenv("POSTGRES_DATABASE_URL", origDSN)
	} else {
		defer os.Unsetenv("POSTGRES_DATABASE_URL")
	}
	os.Unsetenv("POSTGRES_DATABASE_URL")

	origExit := logging.Logger.ExitFunc
	defer func() {
		logging.Logger.ExitFunc = origExit
	}()

	logging.Logger.ExitFunc = func(code int) {
		panic("logger_exit")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic due to Logger.Fatal when POSTGRES_DATABASE_URL is empty, got none")
		}
	}()

	_ = NewPool(context.Background())
}

func TestNewPool_InvalidDSN_TriggersFatal(t *testing.T) {
	origDSN, hadDSN := os.LookupEnv("POSTGRES_DATABASE_URL")
	if hadDSN {
		defer os.Setenv("POSTGRES_DATABASE_URL", origDSN)
	} else {
		defer os.Unsetenv("POSTGRES_DATABASE_URL")
	}
	if err := os.Setenv("POSTGRES_DATABASE_URL", "not-a-valid-dsn"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}

	origExit := logging.Logger.ExitFunc
	defer func() {
		logging.Logger.ExitFunc = origExit
	}()

	logging.Logger.ExitFunc = func(code int) {
		panic("logger_exit")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic due to Logger.Fatal when DSN is invalid, got none")
		}
	}()

	_ = NewPool(context.Background())
}

func TestNewPool_Success_UsesValidEnv(t *testing.T) {
	dsn, ok := os.LookupEnv("POSTGRES_DATABASE_URL")
	if !ok || dsn == "" {
		t.Skip("POSTGRES_DATABASE_URL is not set; skipping integration test")
	}

	origExit := logging.Logger.ExitFunc
	defer func() {
		logging.Logger.ExitFunc = origExit
	}()
	logging.Logger.ExitFunc = func(code int) {
		panic("logger_exit")
	}

	ctx := context.Background()
	pool := NewPool(ctx)
	if pool == nil {
		t.Fatalf("expected non-nil pool")
	}
	pool.Close()
}


