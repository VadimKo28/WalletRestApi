package db

import (
	"context"
	"errors"
	"fmt"

	"walet_rest_api/internal/domain/wallet"
	"walet_rest_api/pkg/client/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type WalletDB struct {
	client postgres.Client
	logger *logrus.Logger
}

func NewWalletDB(client postgres.Client, logger *logrus.Logger) wallet.Storage {
	return &WalletDB{client: client, logger: logger}
}

func (w *WalletDB) ChangeBalance(dto *wallet.WalletChangeBalanceDTO) (*wallet.Wallet, error) {
	ctx := context.Background()

	var query string
	var err error

	switch dto.OperationType {
	case "DEPOSIT":
		// Пополнение баланса
		query = `UPDATE wallets SET balance = balance + $1 WHERE id = $2 RETURNING id, balance`
		w.logger.Info(fmt.Sprintf("SQL query: %s, amount: %d, walletID: %v", query, dto.Balance, dto.ID))

		var wallet wallet.Wallet
		err = w.client.QueryRow(ctx, query, dto.Balance, dto.ID).Scan(&wallet.ID, &wallet.Balance)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("wallet with id %v not found", dto.ID)
			}
			return nil, fmt.Errorf("failed to deposit balance: %w", err)
		}

		return &wallet, nil

	case "WITHDRAW":
		// Списание с баланса (проверяем, что баланс достаточен)
		query = `UPDATE wallets SET balance = balance - $1 WHERE id = $2 AND balance >= $1 RETURNING id, balance`
		w.logger.Info(fmt.Sprintf("SQL query: %s, amount: %d, walletID: %v", query, dto.Balance, dto.ID))

		var wallet wallet.Wallet
		err = w.client.QueryRow(ctx, query, dto.Balance, dto.ID).Scan(&wallet.ID, &wallet.Balance)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// Проверяем, существует ли кошелек
				var exists bool
				checkQuery := `SELECT EXISTS(SELECT 1 FROM wallets WHERE id = $1)`
				if err := w.client.QueryRow(ctx, checkQuery, dto.ID).Scan(&exists); err != nil {
					return nil, fmt.Errorf("failed to check wallet existence: %w", err)
				}
				if !exists {
					return nil, fmt.Errorf("wallet with id %v not found", dto.ID)
				}
				return nil, fmt.Errorf("insufficient balance for withdrawal")
			}
			return nil, fmt.Errorf("failed to withdraw balance: %w", err)
		}

		return &wallet, nil

	default:
		return nil, fmt.Errorf("invalid operation type: %s. Expected DEPOSIT or WITHDRAW", dto.OperationType)
	}
}

func (w *WalletDB) GetBalance(ctx context.Context, walletID string) (int, error) {
	query := `SELECT balance FROM wallets WHERE id = $1`

	w.logger.Info(fmt.Sprintf("SQL query: %s", query))

	var balance int
	if err := w.client.QueryRow(ctx, query, walletID).Scan(&balance); err != nil {
		return 0, err
	}

	return balance, nil
}
