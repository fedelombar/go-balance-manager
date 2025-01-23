package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var baseURL string

func TestMain(m *testing.M) {
	baseURL = "http://localhost:8081" // Updated port to match Docker mapping
	time.Sleep(5 * time.Second)       // Wait for the application to initialize

	code := m.Run()

	os.Exit(code)
}

func TestTransactionAndBalance(t *testing.T) {
	userID := uint64(1)

	initialBalance := getBalance(t, userID)
	assert.Equal(t, "0.00", initialBalance)

	transactionID := generateUniqueTransactionID("txn-integration-win")
	sendTransaction(t, userID, "win", "10.00", transactionID)

	balance := getBalance(t, userID)
	assert.Equal(t, "10.00", balance)

	transactionID = generateUniqueTransactionID("txn-integration-lose")
	sendTransaction(t, userID, "lose", "5.00", transactionID)

	balance = getBalance(t, userID)
	assert.Equal(t, "5.00", balance)

	transactionID = generateUniqueTransactionID("txn-integration-negative")
	resp, err := sendTransactionRaw(t, userID, "lose", "10.00", transactionID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	balance = getBalance(t, userID)
	assert.Equal(t, "5.00", balance)

	duplicateTransactionID := generateUniqueTransactionID("txn-integration-duplicate")

	sendTransaction(t, userID, "win", "10.00", duplicateTransactionID)

	resp, err = sendTransactionRaw(t, userID, "win", "10.00", duplicateTransactionID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	var respData map[string]string
	err = json.Unmarshal(body, &respData)
	assert.NoError(t, err)
	assert.Equal(t, "already processed", respData["status"])

	balance = getBalance(t, userID)
	assert.Equal(t, "15.00", balance)
}

func getBalance(t *testing.T, userID uint64) string {
	resp, err := http.Get(baseURL + "/user/" + strconv.FormatUint(userID, 10) + "/balance")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	var balanceResp BalanceResponse
	err = json.Unmarshal(body, &balanceResp)
	assert.NoError(t, err)

	return balanceResp.Balance
}

func sendTransaction(t *testing.T, userID uint64, state, amount, transactionID string) {
	resp, err := sendTransactionRaw(t, userID, state, amount, transactionID)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	var respData map[string]string
	err = json.Unmarshal(body, &respData)
	assert.NoError(t, err)
	assert.Equal(t, "success", respData["status"])
}

func sendTransactionRaw(t *testing.T, userID uint64, state, amount, transactionID string) (*http.Response, error) {
	transaction := TransactionRequest{
		State:         state,
		Amount:        amount,
		TransactionID: transactionID,
	}

	body, err := json.Marshal(transaction)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/user/"+strconv.FormatUint(userID, 10)+"/transaction", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Source-Type", "game")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(req)
}

func generateUniqueTransactionID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
