package main

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestEnsurePredefinedUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{Db: db}

	for _, id := range []uint64{1, 2, 3} {
		mock.ExpectExec("INSERT INTO users").
			WithArgs(id). // Solo un argumento: user_id
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	err = store.EnsurePredefinedUsers()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUserBalance_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{Db: db}

	userID := uint64(1)
	delta := 10.0
	currentBalance := 50.0
	newBalance := currentBalance + delta

	// init transaction
	mock.ExpectBegin()

	mock.ExpectQuery("SELECT balance FROM users WHERE user_id = \\$1 FOR UPDATE").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(currentBalance))

	// update balance
	mock.ExpectExec("UPDATE users SET balance = \\$1 WHERE user_id = \\$2").
		WithArgs(newBalance, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = store.UpdateUserBalance(userID, delta)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUserBalance_NegativeBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{Db: db}

	userID := uint64(1)
	delta := -60.0
	currentBalance := 50.0

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT balance FROM users WHERE user_id = \\$1 FOR UPDATE").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(currentBalance))

	mock.ExpectRollback()

	err = store.UpdateUserBalance(userID, delta)
	assert.Error(t, err)
	assert.Equal(t, "balance cannot be negative", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserBalance_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{Db: db}
	userID := uint64(1)
	expectedBalance := 100.0

	mock.ExpectQuery("SELECT balance FROM users WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(expectedBalance))

	balance, err := store.GetUserBalance(userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserBalance_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{Db: db}

	userID := uint64(999) // Usuario inexistente

	mock.ExpectQuery("SELECT balance FROM users WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	balance, err := store.GetUserBalance(userID)
	assert.Error(t, err)
	assert.Equal(t, 0.0, balance)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTransaction_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{Db: db}

	tx := Transaction{
		TransactionID: "txn-123",
		UserID:        1,
		State:         "win",
		Amount:        10.0,
		SourceType:    "game",
		CreatedAt:     time.Now(),
	}

	mock.ExpectExec("INSERT INTO transactions").
		WithArgs(tx.TransactionID, tx.UserID, tx.State, tx.Amount, tx.SourceType, tx.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.CreateTransaction(tx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTransaction_Duplicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{Db: db}

	tx := Transaction{
		TransactionID: "txn-duplicate",
		UserID:        1,
		State:         "win",
		Amount:        10.0,
		SourceType:    "game",
		CreatedAt:     time.Now(),
	}

	mock.ExpectExec("INSERT INTO transactions").
		WithArgs(tx.TransactionID, tx.UserID, tx.State, tx.Amount, tx.SourceType, tx.CreatedAt).
		WillReturnError(errors.New("duplicate key value violates unique constraint"))

	err = store.CreateTransaction(tx)
	assert.Error(t, err)
	assert.Equal(t, "duplicate key value violates unique constraint", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}
