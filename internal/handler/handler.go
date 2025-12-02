package handler

import (
	"fmt"
	"net/http"
	"strings"
	"walet_rest_api/internal/domain/wallet"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	walletByUUIDUrl     = "/api/v1/wallets/:wallet_uuid"
	walletChangeBalance = "/api/v1/wallet"
)

type handlers struct {
	service wallet.Service
	logger  *logrus.Logger
}

func NewHandlers(service wallet.Service, logger *logrus.Logger) *handlers {
	return &handlers{service: service, logger: logger}
}

type changeBalanceRequest struct {
	WalletID      uuid.UUID `json:"walletId" binding:"required"`
	OperationType string `json:"operationType" binding:"required"`
	Amount        int    `json:"amount" binding:"required,gt=0"`
}

func (h *handlers) ChangeBalanceWallet(c *gin.Context) {
	var req changeBalanceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(fmt.Sprintf("Invalid request body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.OperationType = strings.ToUpper(req.OperationType)
	if req.OperationType != "DEPOSIT" && req.OperationType != "WITHDRAW" {
		h.logger.Warn(fmt.Sprintf("Invalid operation type: %s", req.OperationType))
		c.JSON(http.StatusBadRequest, gin.H{"error": "operationType must be DEPOSIT or WITHDRAW"})
		return
	}

	dto := &wallet.WalletChangeBalanceDTO{
		ID:            req.WalletID,
		OperationType: req.OperationType,
		Balance:       req.Amount,
	}

	updatedWallet, err := h.service.ChangeBalanceWallet(c.Request.Context(), dto)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to change wallet balance: %v", err))

		// Определяем статус код в зависимости от типа ошибки
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "insufficient balance") {
			statusCode = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "invalid operation type") {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info(fmt.Sprintf("Successfully changed wallet balance for wallet %v", updatedWallet.ID))
	c.JSON(http.StatusOK, gin.H{
		"wallet_id": updatedWallet.ID,
		"balance":   updatedWallet.Balance,
	})
}

func (h *handlers) GetWalletByUUID(c *gin.Context) {
	walletUUID := c.Param("wallet_uuid")

	if walletUUID == "" {
		h.logger.Warn("wallet_uuid parameter is empty")
		c.JSON(400, gin.H{"error": "wallet_uuid is required"})
		return
	}

	balance, err := h.service.GetBalanceWalletByWalletID(c.Request.Context(), walletUUID)
	if err != nil {
		h.logger.Error("Failed to get wallet balance")
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Successfully retrieved wallet balance")

	c.JSON(200, gin.H{"balance": balance})
}

func (h *handlers) RegisterRoutes(router *gin.Engine) {
	router.GET(walletByUUIDUrl, h.GetWalletByUUID)
	router.POST(walletChangeBalance, h.ChangeBalanceWallet)
}
