package models

import "time"

// User represents a user in the banking system
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Exclude password from JSON output
	CreatedAt time.Time `json:"created_at"`
}

// Account represents a bank account
type Account struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	AccountNumber string    `json:"account_number"`
	Balance      float64   `json:"balance"`
	Currency     string    `json:"currency"`
	CreatedAt    time.Time `json:"created_at"`
}

// DepositRequest represents the request body for a deposit
type DepositRequest struct {
	AccountNumber string  `json:"account_number"`
	Amount        float64 `json:"amount"`
}

// WithdrawRequest represents the request body for a withdrawal
type WithdrawRequest struct {
	AccountNumber string  `json:"account_number"`
	Amount        float64 `json:"amount"`
}

// TransferRequest represents the request body for a transfer
type TransferRequest struct {
	FromAccountNumber string  `json:"from_account_number"`
	ToAccountNumber   string  `json:"to_account_number"`
	Amount            float64 `json:"amount"`
}

// CreateUserRequest represents the request body for creating a new user
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateAccountRequest represents the request body for creating a new account
type CreateAccountRequest struct {
	UserID   int    `json:"user_id"`
	Currency string `json:"currency"`
}
