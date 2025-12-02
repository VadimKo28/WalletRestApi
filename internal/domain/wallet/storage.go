package wallet

import "context"

type Storage interface {
	ChangeBalance(dto *WalletChangeBalanceDTO) (*Wallet, error)
	GetBalance(ctx context.Context, walletID string) (int, error)
}