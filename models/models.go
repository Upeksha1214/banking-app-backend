package models

import "time"

// Branch represents a bank branch
type Branch struct {
	BranchID   int    `json:"branch_id"`
	Name       string `json:"name"`
	Location   string `json:"location"`
	ManagerID  *int   `json:"manager_id"` // Use pointer for nullable FK
}

// Employee represents a bank employee
type Employee struct {
	EmployeeID int    `json:"employee_id"`
	Name       string `json:"name"`
	Position   string `json:"position"`
	BranchID   int    `json:"branch_id"`
	ManagerID  *int   `json:"manager_id"` // Use pointer for nullable FK
}

// Customer represents a bank customer
type Customer struct {
	CustomerID int       `json:"customer_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Address    string    `json:"address"`
	DOB        time.Time `json:"dob"` // Date of Birth
	NationalID string    `json:"national_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// User represents a user for login (can be customer, employee, or admin)
type User struct {
	UserID     int    `json:"user_id"`
	Username   string `json:"username"`
	Password   string `json:"-"` // Exclude password from JSON output
	Role       string `json:"role"` // 'admin', 'employee', 'customer'
	CustomerID *int   `json:"customer_id"` // Nullable FK to customers
	EmployeeID *int   `json:"employee_id"` // Nullable FK to employees
	CreatedAt  time.Time `json:"created_at"`
}

// Account represents a bank account (updated fields)
type Account struct {
	AccountID     int       `json:"account_id"`
	CustomerID    int       `json:"customer_id"`
	AccountNumber string    `json:"account_number"`
	AccountType   string    `json:"account_type"` // 'savings', 'current'
	Balance       float64   `json:"balance"`
	OpenedDate    time.Time `json:"opened_date"` // Use time.Time for DATE type
	BranchID      int       `json:"branch_id"`
}

// Transaction represents a financial transaction
type Transaction struct {
	TransactionID   int       `json:"transaction_id"`
	AccountID       int       `json:"account_id"`
	Type            string    `json:"type"` // 'deposit', 'withdrawal', 'transfer_in', 'transfer_out'
	Amount          float64   `json:"amount"`
	TransactionDate time.Time `json:"transaction_date"`
	Description     string    `json:"description"`
}

// Loan represents a loan taken by a customer
type Loan struct {
	LoanID      int       `json:"loan_id"`
	CustomerID  int       `json:"customer_id"`
	Amount      float64   `json:"amount"`
	InterestRate float64   `json:"interest_rate"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Status      string    `json:"status"` // 'approved', 'pending', 'rejected'
}

// Card represents a bank card
type Card struct {
	CardID      int       `json:"card_id"`
	AccountID   int       `json:"account_id"`
	CardNumber  string    `json:"card_number"`
	CardType    string    `json:"card_type"` // 'debit', 'credit'
	ExpiryDate  time.Time `json:"expiry_date"` // Use time.Time for DATE type
	CVV         string    `json:"-"` // Exclude CVV from JSON output, store hashed/encrypted
	CreatedAt   time.Time `json:"created_at"`
}

// --- Request Payloads (Keep existing and add new ones) ---

// CreateUserRequest (updated to include role and optional customer/employee IDs)
type CreateUserRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Role       string `json:"role"` // 'admin', 'employee', 'customer'
	CustomerID *int   `json:"customer_id"` // Use pointer for optional fields
	EmployeeID *int   `json:"employee_id"` // Use pointer for optional fields
}

// CreateCustomerRequest
type CreateCustomerRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
	DOB        string `json:"dob"` // Send as string "YYYY-MM-DD"
	NationalID string `json:"national_id"`
}

// CreateBranchRequest
type CreateBranchRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	ManagerID *int  `json:"manager_id"` // Optional
}

// CreateEmployeeRequest
type CreateEmployeeRequest struct {
	Name      string `json:"name"`
	Position  string `json:"position"`
	BranchID  int    `json:"branch_id"`
	ManagerID *int   `json:"manager_id"` // Optional
}

// CreateAccountRequest (updated)
type CreateAccountRequest struct {
	CustomerID    int    `json:"customer_id"`
	AccountType   string `json:"account_type"` // 'savings', 'current'
	OpenedDate    string `json:"opened_date"`  // Send as string "YYYY-MM-DD"
	BranchID      int    `json:"branch_id"`
}

// DepositRequest (remains the same)
type DepositRequest struct {
	AccountNumber string  `json:"account_number"`
	Amount        float64 `json:"amount"`
}

// WithdrawRequest (remains the same)
type WithdrawRequest struct {
	AccountNumber string  `json:"account_number"`
	Amount        float64 `json:"amount"`
}

// TransferRequest (remains the same)
type TransferRequest struct {
	FromAccountNumber string  `json:"from_account_number"`
	ToAccountNumber   string  `json:"to_account_number"`
	Amount            float64 `json:"amount"`
}

// CreateLoanRequest
type CreateLoanRequest struct {
	CustomerID  int     `json:"customer_id"`
	Amount      float64 `json:"amount"`
	InterestRate float64 `json:"interest_rate"`
	StartDate   string  `json:"start_date"` // Send as string "YYYY-MM-DD"
	EndDate     string  `json:"end_date"`   // Send as string "YYYY-MM-DD"
	Status      string  `json:"status"`     // 'pending' initially
}

// UpdateLoanStatusRequest
type UpdateLoanStatusRequest struct {
	Status string `json:"status"` // 'approved', 'rejected'
}

// CreateCardRequest
type CreateCardRequest struct {
	AccountID  int    `json:"account_id"`
	CardType   string `json:"card_type"` // 'debit', 'credit'
	ExpiryDate string `json:"expiry_date"` // Send as string "YYYY-MM-DD"
	CVV        string `json:"cvv"` // In production, handle securely
}
