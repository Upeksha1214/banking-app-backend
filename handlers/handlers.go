package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"banking-app/db"    // Import our db package
	"banking-app/models" // Import our models package

	"github.com/gorilla/mux"
)

// Helper function to send JSON responses
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Helper function to send error responses
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// GenerateAccountNumber generates a unique 10-digit account number
func GenerateAccountNumber() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%010d", rand.Intn(10000000000)) // 10-digit number
}

// CreateUser handles the creation of a new user
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// In a real app, hash the password here before storing
	// For simplicity, we're storing it as plain text (DO NOT DO THIS IN PRODUCTION)
	result, err := db.DB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", req.Username, req.Password)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID for user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to get user ID")
		return
	}

	user := models.User{
		ID:       int(userID),
		Username: req.Username,
		// Password is not returned
		CreatedAt: time.Now(), // This might be slightly off from DB's timestamp
	}
	respondWithJSON(w, http.StatusCreated, user)
}

// GetUserByID retrieves a user by their ID
func GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user models.User
	err = db.DB.QueryRow("SELECT id, username, created_at FROM users WHERE id = ?", id).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			log.Printf("Error getting user by ID: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve user")
		}
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}

// CreateAccount handles the creation of a new account for a user
func CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Check if user exists
	var userExists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", req.UserID).Scan(&userExists)
	if err != nil || !userExists {
		respondWithError(w, http.StatusBadRequest, "User does not exist")
		return
	}

	accountNumber := GenerateAccountNumber()
	result, err := db.DB.Exec("INSERT INTO accounts (user_id, account_number, balance, currency) VALUES (?, ?, ?, ?)",
		req.UserID, accountNumber, 0.00, req.Currency)
	if err != nil {
		log.Printf("Error creating account: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	accountID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID for account: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to get account ID")
		return
	}

	account := models.Account{
		ID:            int(accountID),
		UserID:        req.UserID,
		AccountNumber: accountNumber,
		Balance:       0.00,
		Currency:      req.Currency,
		CreatedAt:     time.Now(), // This might be slightly off from DB's timestamp
	}
	respondWithJSON(w, http.StatusCreated, account)
}

// GetAccountByNumber retrieves an account by its account number
func GetAccountByNumber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountNumber := vars["accountNumber"]

	var account models.Account
	err := db.DB.QueryRow("SELECT id, user_id, account_number, balance, currency, created_at FROM accounts WHERE account_number = ?", accountNumber).Scan(
		&account.ID, &account.UserID, &account.AccountNumber, &account.Balance, &account.Currency, &account.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Account not found")
		} else {
			log.Printf("Error getting account by number: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve account")
		}
		return
	}
	respondWithJSON(w, http.StatusOK, account)
}

// Deposit funds into an account
func Deposit(w http.ResponseWriter, r *http.Request) {
	var req models.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Amount <= 0 {
		respondWithError(w, http.StatusBadRequest, "Deposit amount must be positive")
		return
	}

	// Start a transaction for atomicity
	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction for deposit: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to process deposit")
		return
	}
	defer tx.Rollback() // Rollback on error, commit if successful

	// Get current balance with a FOR UPDATE lock
	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM accounts WHERE account_number = ? FOR UPDATE", req.AccountNumber).Scan(&currentBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Account not found")
		} else {
			log.Printf("Error fetching account for deposit: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve account balance")
		}
		return
	}

	newBalance := currentBalance + req.Amount
	_, err = tx.Exec("UPDATE accounts SET balance = ? WHERE account_number = ?", newBalance, req.AccountNumber)
	if err != nil {
		log.Printf("Error updating account balance for deposit: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update account balance")
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction for deposit: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to commit deposit transaction")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Deposit successful",
		"account_number": req.AccountNumber,
		"new_balance":  newBalance,
	})
}

// Withdraw funds from an account
func Withdraw(w http.ResponseWriter, r *http.Request) {
	var req models.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Amount <= 0 {
		respondWithError(w, http.StatusBadRequest, "Withdrawal amount must be positive")
		return
	}

	// Start a transaction for atomicity
	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction for withdrawal: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to process withdrawal")
		return
	}
	defer tx.Rollback() // Rollback on error, commit if successful

	// Get current balance with a FOR UPDATE lock
	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM accounts WHERE account_number = ? FOR UPDATE", req.AccountNumber).Scan(&currentBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Account not found")
		} else {
			log.Printf("Error fetching account for withdrawal: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve account balance")
		}
		return
	}

	if currentBalance < req.Amount {
		respondWithError(w, http.StatusBadRequest, "Insufficient funds")
		return
	}

	newBalance := currentBalance - req.Amount
	_, err = tx.Exec("UPDATE accounts SET balance = ? WHERE account_number = ?", newBalance, req.AccountNumber)
	if err != nil {
		log.Printf("Error updating account balance for withdrawal: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update account balance")
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction for withdrawal: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to commit withdrawal transaction")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Withdrawal successful",
		"account_number": req.AccountNumber,
		"new_balance":  newBalance,
	})
}

// Transfer funds between accounts
func Transfer(w http.ResponseWriter, r *http.Request) {
	var req models.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Amount <= 0 {
		respondWithError(w, http.StatusBadRequest, "Transfer amount must be positive")
		return
	}
	if req.FromAccountNumber == req.ToAccountNumber {
		respondWithError(w, http.StatusBadRequest, "Cannot transfer to the same account")
		return
	}

	// Start a transaction for atomicity
	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction for transfer: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to process transfer")
		return
	}
	defer tx.Rollback() // Rollback on error, commit if successful

	// Get balances of both accounts with FOR UPDATE locks
	var fromBalance, toBalance float64
	err = tx.QueryRow("SELECT balance FROM accounts WHERE account_number = ? FOR UPDATE", req.FromAccountNumber).Scan(&fromBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Source account not found")
		} else {
			log.Printf("Error fetching source account for transfer: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve source account balance")
		}
		return
	}

	err = tx.QueryRow("SELECT balance FROM accounts WHERE account_number = ? FOR UPDATE", req.ToAccountNumber).Scan(&toBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Destination account not found")
		} else {
			log.Printf("Error fetching destination account for transfer: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve destination account balance")
		}
		return
	}

	if fromBalance < req.Amount {
		respondWithError(w, http.StatusBadRequest, "Insufficient funds in source account")
		return
	}

	// Update balances
	_, err = tx.Exec("UPDATE accounts SET balance = ? WHERE account_number = ?", fromBalance-req.Amount, req.FromAccountNumber)
	if err != nil {
		log.Printf("Error updating source account balance for transfer: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update source account balance")
		return
	}

	_, err = tx.Exec("UPDATE accounts SET balance = ? WHERE account_number = ?", toBalance+req.Amount, req.ToAccountNumber)
	if err != nil {
		log.Printf("Error updating destination account balance for transfer: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update destination account balance")
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction for transfer: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to commit transfer transaction")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Transfer successful"})
}
