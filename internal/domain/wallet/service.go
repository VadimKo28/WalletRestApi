package wallet

import (
	"context"
)

type Service interface {
	ChangeBalanceWallet(ctx context.Context, dto *WalletChangeBalanceDTO) (*Wallet, error)
	GetBalanceWalletByWalletID(ctx context.Context, walletID string) (int, error)
}

type service struct {
	storage Storage
}

func (s *service) ChangeBalanceWallet(ctx context.Context, dto *WalletChangeBalanceDTO) (*Wallet, error) {
	return s.storage.ChangeBalance(dto)
}

func (s *service) GetBalanceWalletByWalletID(ctx context.Context, walletID string) (int, error) {
	return s.storage.GetBalance(ctx, walletID)
}

func NewService(storage Storage) Service {
	return &service{storage: storage}
}
