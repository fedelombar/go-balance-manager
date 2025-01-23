package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type MockStore struct {
	Transactions map[string]Transaction
	Users        map[uint64]float64
}

func NewMockStore() *MockStore {
	return &MockStore{
		Transactions: make(map[string]Transaction),
		Users:        map[uint64]float64{1: 0.0, 2: 0.0, 3: 0.0},
	}
}

func (m *MockStore) CreateTransaction(tx Transaction) error {
	if _, exists := m.Transactions[tx.TransactionID]; exists {
		return errors.New("duplicate transaction")
	}
	m.Transactions[tx.TransactionID] = tx
	return nil
}

func (m *MockStore) GetUserBalance(userID uint64) (float64, error) {
	balance, exists := m.Users[userID]
	if !exists {
		return 0, errors.New("user not found")
	}
	return balance, nil
}

func (m *MockStore) GetTransactionByID(transactionID string) (*Transaction, error) {
	tx, exists := m.Transactions[transactionID]
	if !exists {
		return nil, errors.New("transaction not found")
	}
	return &tx, nil
}

func (m *MockStore) UpdateUserBalance(userID uint64, delta float64) error {
	current, exists := m.Users[userID]
	if !exists {
		return errors.New("user not found")
	}
	newBalance := current + delta
	if newBalance < 0 {
		return errors.New("balance cannot be negative")
	}
	m.Users[userID] = newBalance
	return nil
}

func (m *MockStore) EnsurePredefinedUsers() error {
	return nil
}

func (m *MockStore) Init() error {
	return nil
}

func (m *MockStore) Close() error {
	return nil
}

func TestHandleTransaction_SuccessWin(t *testing.T) {
	store := NewMockStore()
	server := NewAPIServer(store)

	router := mux.NewRouter()
	router.HandleFunc("/user/{userId}/transaction", server.HandleTransaction).Methods("POST")

	transaction := TransactionRequest{
		State:         "win",
		Amount:        "10.15",
		TransactionID: "txn-abc123",
	}

	body, _ := json.Marshal(transaction)
	req, err := http.NewRequest("POST", "/user/1/transaction", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Source-Type", "game")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp["status"])

	balance, err := store.GetUserBalance(1)
	assert.NoError(t, err)
	assert.Equal(t, 10.15, balance)
}

func TestHandleTransaction_SuccessLose(t *testing.T) {
	store := NewMockStore()
	store.Users[1] = 10.0
	server := NewAPIServer(store)

	router := mux.NewRouter()
	router.HandleFunc("/user/{userId}/transaction", server.HandleTransaction).Methods("POST")

	transaction := TransactionRequest{
		State:         "lose",
		Amount:        "5.00",
		TransactionID: "txn-def456",
	}

	body, _ := json.Marshal(transaction)
	req, err := http.NewRequest("POST", "/user/1/transaction", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Source-Type", "server")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp["status"])

	// verify balance
	balance, err := store.GetUserBalance(1)
	assert.NoError(t, err)
	assert.Equal(t, 5.0, balance)
}

func TestHandleTransaction_Duplicate(t *testing.T) {
	store := NewMockStore()
	server := NewAPIServer(store)

	// create prev transaction
	store.Transactions["txn-abc123"] = Transaction{
		TransactionID: "txn-abc123",
		UserID:        1,
		State:         "win",
		Amount:        10.15,
		SourceType:    "game",
		CreatedAt:     time.Now(),
	}
	store.Users[1] = 10.15

	router := mux.NewRouter()
	router.HandleFunc("/user/{userId}/transaction", server.HandleTransaction).Methods("POST")

	transaction := TransactionRequest{
		State:         "win",
		Amount:        "10.15",
		TransactionID: "txn-abc123",
	}

	body, _ := json.Marshal(transaction)
	req, err := http.NewRequest("POST", "/user/1/transaction", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Source-Type", "game")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "already processed", resp["status"])

	// verify the balance didn't update it
	balance, err := store.GetUserBalance(1)
	assert.NoError(t, err)
	assert.Equal(t, 10.15, balance)
}

func TestHandleTransaction_InvalidState(t *testing.T) {
	store := NewMockStore()
	server := NewAPIServer(store)

	router := mux.NewRouter()
	router.HandleFunc("/user/{userId}/transaction", server.HandleTransaction).Methods("POST")

	transaction := TransactionRequest{
		State:         "invalid_state",
		Amount:        "10.00",
		TransactionID: "txn-invalid",
	}

	body, _ := json.Marshal(transaction)
	req, err := http.NewRequest("POST", "/user/1/transaction", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Source-Type", "game")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid state value")
}

func TestHandleTransaction_InvalidAmount(t *testing.T) {
	store := NewMockStore()
	server := NewAPIServer(store)

	router := mux.NewRouter()
	router.HandleFunc("/user/{userId}/transaction", server.HandleTransaction).Methods("POST")

	// Amount con más de 2 decimales
	transaction := TransactionRequest{
		State:         "win",
		Amount:        "10.123",
		TransactionID: "txn-invalid-amount",
	}

	body, _ := json.Marshal(transaction)
	req, err := http.NewRequest("POST", "/user/1/transaction", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Source-Type", "game")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Amount must have up to 2 decimal places")
}

func TestHandleGetBalance_Success(t *testing.T) {
	store := NewMockStore()
	store.Users[1] = 25.75
	server := NewAPIServer(store)

	router := mux.NewRouter()
	router.HandleFunc("/user/{userId}/balance", server.HandleGetBalance).Methods("GET")

	req, err := http.NewRequest("GET", "/user/1/balance", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp BalanceResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), resp.UserID)
	assert.Equal(t, "25.75", resp.Balance)
}

func TestHandleGetBalance_UserNotFound(t *testing.T) {
	store := NewMockStore()
	server := NewAPIServer(store)

	router := mux.NewRouter()
	router.HandleFunc("/user/{userId}/balance", server.HandleGetBalance).Methods("GET")

	req, err := http.NewRequest("GET", "/user/999/balance", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "User not found")
}
