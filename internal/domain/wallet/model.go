package wallet

import "github.com/google/uuid"


type Wallet struct {
	ID uuid.UUID `json:"wallet_id"`
	Balance int `json:"balance"`
}

type WalletChangeBalanceDTO struct {
	ID uuid.UUID `json:"wallet_id" binding:"required"`
	OperationType string `json:"operationType" binding:"required"`
	Balance int `json:"balance"`
}
