package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateTransaction(tx Transaction) error
	GetUserBalance(userID uint64) (float64, error)
	GetTransactionByID(transactionID string) (*Transaction, error)
	UpdateUserBalance(userID uint64, delta float64) error
	EnsurePredefinedUsers() error
	Init() error
	Close() error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(host string, port int, user, password, dbname string) (*PostgresStore, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

// Init runs database migrations and seeds predefined users
func (s *PostgresStore) Init() error {
	if err := s.createUsersTable(); err != nil {
		return err
	}
	if err := s.createTransactionsTable(); err != nil {
		return err
	}
	return s.EnsurePredefinedUsers()
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) createUsersTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		user_id BIGSERIAL PRIMARY KEY,
		balance NUMERIC(12, 2) NOT NULL DEFAULT 0,
		created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
	);`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) createTransactionsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS transactions (
		id BIGSERIAL PRIMARY KEY,
		transaction_id VARCHAR(255) UNIQUE NOT NULL,
		user_id BIGINT NOT NULL,
		state VARCHAR(10) NOT NULL,   -- "win" or "lose"
		amount NUMERIC(12, 2) NOT NULL,
		source_type VARCHAR(50) NOT NULL,   -- "game", "server", "payment", etc.
		created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
		FOREIGN KEY (user_id) REFERENCES users(user_id)
	);`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateTransaction(tx Transaction) error {
	query := `
	INSERT INTO transactions (transaction_id, user_id, state, amount, source_type, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.db.Exec(query,
		tx.TransactionID,
		tx.UserID,
		tx.State,
		tx.Amount,
		tx.SourceType,
		tx.CreatedAt,
	)
	return err
}

func (s *PostgresStore) GetUserBalance(userID uint64) (float64, error) {
	var balance float64
	err := s.db.QueryRow("SELECT balance FROM users WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (s *PostgresStore) GetTransactionByID(transactionID string) (*Transaction, error) {
	query := `
	SELECT id, transaction_id, user_id, state, amount, source_type, created_at
	FROM transactions WHERE transaction_id = $1`
	row := s.db.QueryRow(query, transactionID)

	var tx Transaction
	err := row.Scan(
		&tx.ID,
		&tx.TransactionID,
		&tx.UserID,
		&tx.State,
		&tx.Amount,
		&tx.SourceType,
		&tx.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// UpdateUserBalance updates the user's balance by a delta
func (s *PostgresStore) UpdateUserBalance(userID uint64, delta float64) error {
	// Use a transaction to ensure atomicity
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Lock the user row for update
	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM users WHERE user_id = $1 FOR UPDATE", userID).Scan(&currentBalance)
	if err != nil {
		return err
	}

	newBalance := currentBalance + delta
	if newBalance < 0 {
		return fmt.Errorf("balance cannot be negative")
	}

	_, err = tx.Exec("UPDATE users SET balance = $1 WHERE user_id = $2", newBalance, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// EnsurePredefinedUsers creates users with IDs 1, 2, and 3 if they don't exist
func (s *PostgresStore) EnsurePredefinedUsers() error {
	for _, id := range []uint64{1, 2, 3} {
		_, err := s.db.Exec(`
			INSERT INTO users (user_id, balance)
			VALUES ($1, 0)
			ON CONFLICT (user_id) DO NOTHING
		`, id)
		if err != nil {
			return err
		}
	}
	return nil
}
