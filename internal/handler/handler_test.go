package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"walet_rest_api/internal/domain/wallet"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type mockWalletService struct {
	ChangeBalanceWalletFunc         func(ctx context.Context, dto *wallet.WalletChangeBalanceDTO) (*wallet.Wallet, error)
	GetBalanceWalletByWalletIDFunc  func(ctx context.Context, walletID string) (int, error)
	LastChangeBalanceWalletDTO      *wallet.WalletChangeBalanceDTO
	LastGetBalanceWalletByWalletID  string
}

func (m *mockWalletService) ChangeBalanceWallet(ctx context.Context, dto *wallet.WalletChangeBalanceDTO) (*wallet.Wallet, error) {
	m.LastChangeBalanceWalletDTO = dto
	if m.ChangeBalanceWalletFunc != nil {
		return m.ChangeBalanceWalletFunc(ctx, dto)
	}
	return nil, nil
}

func (m *mockWalletService) GetBalanceWalletByWalletID(ctx context.Context, walletID string) (int, error) {
	m.LastGetBalanceWalletByWalletID = walletID
	if m.GetBalanceWalletByWalletIDFunc != nil {
		return m.GetBalanceWalletByWalletIDFunc(ctx, walletID)
	}
	return 0, nil
}

func setupTestRouter(t *testing.T, service wallet.Service) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()

	h := NewHandlers(service, logger)
	h.RegisterRoutes(router)

	return router
}

func TestChangeBalanceWallet_Success(t *testing.T) {
	mockService := &mockWalletService{}
	router := setupTestRouter(t, mockService)

	walletID := uuid.New()
	reqBody := changeBalanceRequest{
		WalletID:      walletID,
		OperationType: "deposit",
		Amount:        100,
	}

	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	expectedWallet := &wallet.Wallet{
		ID:      walletID,
		Balance: 200,
	}

	mockService.ChangeBalanceWalletFunc = func(ctx context.Context, dto *wallet.WalletChangeBalanceDTO) (*wallet.Wallet, error) {
		assert.Equal(t, walletID, dto.ID)
		assert.Equal(t, "DEPOSIT", dto.OperationType)
		assert.Equal(t, 100, dto.Balance)
		return expectedWallet, nil
	}

	req := httptest.NewRequest(http.MethodPost, walletChangeBalance, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var respBody map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedWallet.ID.String(), respBody["wallet_id"])
	assert.EqualValues(t, expectedWallet.Balance, respBody["balance"])
}

func TestChangeBalanceWallet_InvalidBody(t *testing.T) {
	mockService := &mockWalletService{}
	router := setupTestRouter(t, mockService)

	body := `{"walletId": "", "operationType": "", "amount": 0}`

	req := httptest.NewRequest(http.MethodPost, walletChangeBalance, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var respBody map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request body", respBody["error"])
}

func TestChangeBalanceWallet_InvalidOperationType(t *testing.T) {
	mockService := &mockWalletService{}
	router := setupTestRouter(t, mockService)

	walletID := uuid.New()
	reqBody := changeBalanceRequest{
		WalletID:      walletID,
		OperationType: "unknown",
		Amount:        100,
	}

	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, walletChangeBalance, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var respBody map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "operationType must be DEPOSIT or WITHDRAW", respBody["error"])
}

func TestChangeBalanceWallet_NotFound(t *testing.T) {
	mockService := &mockWalletService{}
	router := setupTestRouter(t, mockService)

	walletID := uuid.New()
	reqBody := changeBalanceRequest{
		WalletID:      walletID,
		OperationType: "WITHDRAW",
		Amount:        50,
	}

	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	mockService.ChangeBalanceWalletFunc = func(ctx context.Context, dto *wallet.WalletChangeBalanceDTO) (*wallet.Wallet, error) {
		return nil, errors.New("wallet not found")
	}

	req := httptest.NewRequest(http.MethodPost, walletChangeBalance, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var respBody map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "wallet not found", respBody["error"])
}

func TestChangeBalanceWallet_InsufficientBalance(t *testing.T) {
	mockService := &mockWalletService{}
	router := setupTestRouter(t, mockService)

	walletID := uuid.New()
	reqBody := changeBalanceRequest{
		WalletID:      walletID,
		OperationType: "WITHDRAW",
		Amount:        150,
	}

	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	mockService.ChangeBalanceWalletFunc = func(ctx context.Context, dto *wallet.WalletChangeBalanceDTO) (*wallet.Wallet, error) {
		return nil, errors.New("insufficient balance")
	}

	req := httptest.NewRequest(http.MethodPost, walletChangeBalance, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var respBody map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "insufficient balance", respBody["error"])
}

func TestChangeBalanceWallet_InvalidOperationTypeFromService(t *testing.T) {
	mockService := &mockWalletService{}
	router := setupTestRouter(t, mockService)

	walletID := uuid.New()
	reqBody := changeBalanceRequest{
		WalletID:      walletID,
		OperationType: "DEPOSIT",
		Amount:        100,
	}

	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	mockService.ChangeBalanceWalletFunc = func(ctx context.Context, dto *wallet.WalletChangeBalanceDTO) (*wallet.Wallet, error) {
		return nil, errors.New("invalid operation type")
	}

	req := httptest.NewRequest(http.MethodPost, walletChangeBalance, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var respBody map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "invalid operation type", respBody["error"])
}

func TestChangeBalanceWallet_InternalError(t *testing.T) {
	mockService := &mockWalletService{}
	router := setupTestRouter(t, mockService)

	walletID := uuid.New()
	reqBody := changeBalanceRequest{
		WalletID:      walletID,
		OperationType: "DEPOSIT",
		Amount:        100,
	}

	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	mockService.ChangeBalanceWalletFunc = func(ctx context.Context, dto *wallet.WalletChangeBalanceDTO) (*wallet.Wallet, error) {
		return nil, errors.New("some internal error")
	}

	req := httptest.NewRequest(http.MethodPost, walletChangeBalance, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var respBody map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "some internal error", respBody["error"])
}

func TestGetWalletByUUID_Success(t *testing.T) {
	mockService := &mockWalletService{}
	router := setupTestRouter(t, mockService)

	walletUUID := uuid.New().String()

	mockService.GetBalanceWalletByWalletIDFunc = func(ctx context.Context, walletID string) (int, error) {
		assert.Equal(t, walletUUID, walletID)
		return 500, nil
	}

	url := "/api/v1/wallets/" + walletUUID
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var respBody map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.EqualValues(t, 500, respBody["balance"])
}

func TestGetWalletByUUID_EmptyParam(t *testing.T) {
	mockService := &mockWalletService{}
	router := setupTestRouter(t, mockService)

	url := "/api/v1/wallets/"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetWalletByUUID_ServiceError(t *testing.T) {
	mockService := &mockWalletService{}
	router := setupTestRouter(t, mockService)

	walletUUID := uuid.New().String()

	mockService.GetBalanceWalletByWalletIDFunc = func(ctx context.Context, walletID string) (int, error) {
		return 0, errors.New("db error")
	}

	url := "/api/v1/wallets/" + walletUUID
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var respBody map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "db error", respBody["error"])
}

func TestRegisterRoutes(t *testing.T) {
	mockService := &mockWalletService{}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()

	mockService.GetBalanceWalletByWalletIDFunc = func(ctx context.Context, walletID string) (int, error) {
		return 0, nil
	}
	mockService.ChangeBalanceWalletFunc = func(ctx context.Context, dto *wallet.WalletChangeBalanceDTO) (*wallet.Wallet, error) {
		return &wallet.Wallet{
			ID:      uuid.New(),
			Balance: 10,
		}, nil
	}

	h := NewHandlers(mockService, logger)
	h.RegisterRoutes(router)

	// Check that routes are registered and respond with some status (not 404)
	walletUUID := uuid.New().String()

	// GET wallet by UUID
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+walletUUID, nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	assert.NotEqual(t, http.StatusNotFound, getRec.Code)

	// POST change balance
	body := `{"walletId":"` + uuid.New().String() + `","operationType":"DEPOSIT","amount":10}`
	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader([]byte(body)))
	postReq.Header.Set("Content-Type", "application/json")
	postRec := httptest.NewRecorder()
	router.ServeHTTP(postRec, postReq)
	assert.NotEqual(t, http.StatusNotFound, postRec.Code)
}


