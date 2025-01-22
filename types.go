package main

import "time"

type TransactionRequest struct {
	State         string `json:"state"` // "win" or "lose"
	Amount        string `json:"amount"`
	TransactionID string `json:"transactionId"`
}

type BalanceResponse struct {
	UserID  uint64 `json:"userId"`
	Balance string `json:"balance"`
}

type User struct {
	UserID  uint64    `json:"userId"`
	Balance float64   `json:"balance"`
	Created time.Time `json:"created_at"`
}

type Transaction struct {
	ID            int64     `json:"id"`
	TransactionID string    `json:"transactionId"`
	UserID        uint64    `json:"userId"`
	State         string    `json:"state"`
	Amount        float64   `json:"amount"`
	SourceType    string    `json:"sourceType"`
	CreatedAt     time.Time `json:"created_at"`
}
