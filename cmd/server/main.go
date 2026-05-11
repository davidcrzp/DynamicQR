package main

import (
	"fmt"
	"log"
	"github.com/davidcrzp/dynamicqr/internal/api"
	"github.com/davidcrzp/dynamicqr/internal/database"
	"github.com/davidcrzp/dynamicqr/internal/qr"
)

func main() {
	db, err := database.NewDB("records.db")
    if err != nil {
        log.Fatalf("Failed to connect to DB: %v", err)
    }
    defer db.Close()

	db.RegisterDoctor("licence123", "Dr. John Doe", "hospital@gmail.com", "password123")
	db.RegisterPatient(1, "John Sick", "patient@gmail.com", "(33) 1122-3344")
	
	_, err = qr.QRGen()
	if err != nil {
		log.Fatalf("Failed to generate QR: %v", err)
	}

	srv := api.NewServer()

    fmt.Println("Server starting on :8080")
	if err := api.StartServer(srv, ":8080"); err != nil {
		log.Fatal(err)
	}
}

