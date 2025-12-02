package db

import (
	"context"
	"errors"
	"strings"
	"testing"

	"walet_rest_api/internal/domain/wallet"
	"walet_rest_api/pkg/client/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"
)

type mockRow struct {
	scanFunc func(dest ...any) error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.scanFunc != nil {
		return m.scanFunc(dest...)
	}
	return nil
}

type mockClient struct {
	execFunc     func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	queryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (m *mockClient) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, sql, arguments...)
	}
	var tag pgconn.CommandTag
	return tag, nil
}

func (m *mockClient) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(ctx, sql, args...)
	}
	return &mockRow{}
}

func newTestWalletDB(t *testing.T, client postgres.Client) *WalletDB {
	t.Helper()
	logger := logrus.New()
	return &WalletDB{
		client: client,
		logger: logger,
	}
}

func TestWalletDB_ChangeBalance_Deposit_Success(t *testing.T) {
	ctx := context.Background()
	walletID := uuid.New()

	client := &mockClient{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if strings.HasPrefix(sql, "SELECT EXISTS") {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						// dest[0] *bool
						if len(dest) != 1 {
							t.Fatalf("expected 1 dest for walletExists, got %d", len(dest))
						}
						ptr, ok := dest[0].(*bool)
						if !ok {
							t.Fatalf("expected *bool for walletExists scan dest")
						}
						*ptr = true
						return nil
					},
				}
			}

			return &mockRow{
				scanFunc: func(dest ...any) error {
					if len(dest) != 2 {
						t.Fatalf("expected 2 dests for update scan, got %d", len(dest))
					}
					idPtr, ok := dest[0].(*uuid.UUID)
					if !ok {
						t.Fatalf("expected *uuid.UUID for first dest")
					}
					balancePtr, ok := dest[1].(*int)
					if !ok {
						t.Fatalf("expected *int for second dest")
					}
					*idPtr = walletID
					*balancePtr = 200
					return nil
				},
			}
		},
	}

	storage := newTestWalletDB(t, client)

	dto := &wallet.WalletChangeBalanceDTO{
		ID:            walletID,
		OperationType: "DEPOSIT",
		Balance:       100,
	}

	w, err := storage.ChangeBalance(ctx, dto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w == nil {
		t.Fatalf("expected non-nil wallet")
	}
	if w.ID != walletID {
		t.Errorf("expected ID %v, got %v", walletID, w.ID)
	}
	if w.Balance != 200 {
		t.Errorf("expected balance 200, got %d", w.Balance)
	}
}

func TestWalletDB_ChangeBalance_Withdraw_InsufficientBalance(t *testing.T) {
	ctx := context.Background()
	walletID := uuid.New()

	client := &mockClient{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if strings.HasPrefix(sql, "SELECT EXISTS") {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						ptr := dest[0].(*bool)
						*ptr = true
						return nil
					},
				}
			}

			return &mockRow{
				scanFunc: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			}
		},
	}

	storage := newTestWalletDB(t, client)

	dto := &wallet.WalletChangeBalanceDTO{
		ID:            walletID,
		OperationType: "WITHDRAW",
		Balance:       500,
	}

	_, err := storage.ChangeBalance(ctx, dto)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "insufficient balance") {
		t.Errorf("expected insufficient balance error, got %v", err)
	}
}

func TestWalletDB_ChangeBalance_WalletDoesNotExist(t *testing.T) {
	ctx := context.Background()
	walletID := uuid.New()

	client := &mockClient{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			// walletExists -> false
			return &mockRow{
				scanFunc: func(dest ...any) error {
					ptr := dest[0].(*bool)
					*ptr = false
					return nil
				},
			}
		},
	}

	storage := newTestWalletDB(t, client)

	dto := &wallet.WalletChangeBalanceDTO{
		ID:            walletID,
		OperationType: "DEPOSIT",
		Balance:       100,
	}

	_, err := storage.ChangeBalance(ctx, dto)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestWalletDB_ChangeBalance_InvalidOperationType(t *testing.T) {
	ctx := context.Background()
	walletID := uuid.New()

	client := &mockClient{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if strings.HasPrefix(sql, "SELECT EXISTS") {
				return &mockRow{
					scanFunc: func(dest ...any) error {
						if len(dest) != 1 {
							t.Fatalf("expected 1 dest for walletExists, got %d", len(dest))
						}
						ptr, ok := dest[0].(*bool)
						if !ok {
							t.Fatalf("expected *bool for walletExists scan dest")
						}
						*ptr = true
						return nil
					},
				}
			}
			t.Fatalf("unexpected sql in invalid operation type test: %s", sql)
			return &mockRow{}
		},
	}

	storage := newTestWalletDB(t, client)

	dto := &wallet.WalletChangeBalanceDTO{
		ID:            walletID,
		OperationType: "UNKNOWN",
		Balance:       100,
	}

	_, err := storage.ChangeBalance(ctx, dto)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid operation type") {
		t.Errorf("expected invalid operation type error, got %v", err)
	}
}

func TestWalletDB_ChangeBalance_ExistsCheckError(t *testing.T) {
	ctx := context.Background()
	walletID := uuid.New()

	client := &mockClient{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFunc: func(dest ...any) error {
					return errors.New("db error")
				},
			}
		},
	}

	storage := newTestWalletDB(t, client)

	dto := &wallet.WalletChangeBalanceDTO{
		ID:            walletID,
		OperationType: "DEPOSIT",
		Balance:       100,
	}

	_, err := storage.ChangeBalance(ctx, dto)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to check wallet existence") {
		t.Errorf("expected wrapped existence check error, got %v", err)
	}
}

func TestWalletDB_GetBalance_Success(t *testing.T) {
	ctx := context.Background()

	client := &mockClient{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			if !strings.HasPrefix(sql, "SELECT balance") {
				t.Fatalf("unexpected sql: %s", sql)
			}
			return &mockRow{
				scanFunc: func(dest ...any) error {
					if len(dest) != 1 {
						t.Fatalf("expected 1 dest for balance scan, got %d", len(dest))
					}
					balancePtr, ok := dest[0].(*int)
					if !ok {
						t.Fatalf("expected *int for balance dest")
					}
					*balancePtr = 300
					return nil
				},
			}
		},
	}

	storage := newTestWalletDB(t, client)

	balance, err := storage.GetBalance(ctx, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance != 300 {
		t.Errorf("expected balance 300, got %d", balance)
	}
}

func TestWalletDB_GetBalance_Error(t *testing.T) {
	ctx := context.Background()

	client := &mockClient{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFunc: func(dest ...any) error {
					return errors.New("db error")
				},
			}
		},
	}

	storage := newTestWalletDB(t, client)

	_, err := storage.GetBalance(ctx, uuid.New().String())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "db error") {
		t.Errorf("expected db error, got %v", err)
	}
}


