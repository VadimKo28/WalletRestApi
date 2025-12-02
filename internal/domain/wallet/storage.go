package wallet

import "context"

type Storage interface {
	ChangeBalance(ctx context.Context, dto *WalletChangeBalanceDTO) (*Wallet, error)
	GetBalance(ctx context.Context, walletID string) (int, error)
}