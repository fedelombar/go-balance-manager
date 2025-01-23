package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type APIServer struct {
	store Storage
}

func NewAPIServer(store Storage) *APIServer {
	return &APIServer{
		store: store,
	}
}

func (s *APIServer) Run(addr string) {
	router := mux.NewRouter()

	router.HandleFunc("/user/{userId}/transaction", s.HandleTransaction).Methods("POST")
	router.HandleFunc("/user/{userId}/balance", s.HandleGetBalance).Methods("GET")
	router.HandleFunc("/health", s.HandleHealth).Methods("GET")

	log.Printf("Server running on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func (s *APIServer) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// HandleTransaction processes POST /user/{userId}/transaction
func (s *APIServer) HandleTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	sourceType := r.Header.Get("Source-Type")
	if sourceType == "" {
		http.Error(w, "Missing Source-Type header", http.StatusBadRequest)
		return
	}

	// Parse JSON body
	var txReq TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&txReq); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate state
	if txReq.State != "win" && txReq.State != "lose" {
		http.Error(w, "Invalid state value", http.StatusBadRequest)
		return
	}

	// Validate amount
	amount, err := parseAmount(txReq.Amount)
	if err != nil {
		http.Error(w, "Invalid amount format", http.StatusBadRequest)
		return
	}

	if !isValidAmount(txReq.Amount) {
		http.Error(w, "Amount must have up to 2 decimal places", http.StatusBadRequest)
		return
	}

	// Check if transaction ID already exists
	existingTx, err := s.store.GetTransactionByID(txReq.TransactionID)
	if err == nil && existingTx != nil {
		// Transaction already processed
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "already processed",
		})
		return
	}

	// Determine balance delta
	var delta float64
	if txReq.State == "win" {
		delta = amount
	} else {
		delta = -amount
	}

	err = s.store.UpdateUserBalance(userID, delta)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Record the transaction
	tx := Transaction{
		TransactionID: txReq.TransactionID,
		UserID:        userID,
		State:         txReq.State,
		Amount:        amount,
		SourceType:    sourceType,
		CreatedAt:     timeNowUTC(),
	}
	if err := s.store.CreateTransaction(tx); err != nil {
		http.Error(w, "Failed to record transaction", http.StatusInternalServerError)
		return
	}

	// res with success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func (s *APIServer) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	balance, err := s.store.GetUserBalance(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	resp := BalanceResponse{
		UserID:  userID,
		Balance: formatAmount(balance),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func parseAmount(amountStr string) (float64, error) {
	return strconv.ParseFloat(amountStr, 64)
}

func isValidAmount(amountStr string) bool {
	if _, err := strconv.ParseFloat(amountStr, 64); err != nil {
		return false
	}
	parts := split(amountStr, '.')
	if len(parts) == 2 && len(parts[1]) > 2 {
		return false
	}
	return true
}

func split(s string, sep rune) []string {
	var parts []string
	current := ""
	for _, c := range s {
		if c == sep {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	parts = append(parts, current)
	return parts
}

func formatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

func timeNowUTC() time.Time {
	return time.Now().UTC()
}
