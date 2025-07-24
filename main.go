package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"banking-app/db"
	"banking-app/handlers"

	"github.com/gorilla/mux"
)

func main() {
	// Database connection string (replace with your MySQL credentials)
	// Example: "user:password@tcp(127.0.0.1:3306)/banking_app"
	// It's best practice to get this from environment variables or a config file
	dataSourceName := os.Getenv("MYSQL_DSN")
	if dataSourceName == "" {
		log.Fatal("MYSQL_DSN environment variable not set. Please set it to your MySQL connection string.")
	}

	// Initialize the database connection
	db.InitDB(dataSourceName)
	defer db.CloseDB() // Ensure database connection is closed when main exits

	// Create a new Gorilla Mux router
	router := mux.NewRouter()

	// User routes
	router.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	router.HandleFunc("/users/{id}", handlers.GetUserByID).Methods("GET")

	// Account routes
	router.HandleFunc("/accounts", handlers.CreateAccount).Methods("POST")
	router.HandleFunc("/accounts/{accountNumber}", handlers.GetAccountByNumber).Methods("GET")

	// Transaction routes
	router.HandleFunc("/accounts/deposit", handlers.Deposit).Methods("POST")
	router.HandleFunc("/accounts/withdraw", handlers.Withdraw).Methods("POST")
	router.HandleFunc("/accounts/transfer", handlers.Transfer).Methods("POST")

	// Start the HTTP server
	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}
