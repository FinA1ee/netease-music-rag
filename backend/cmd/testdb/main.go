package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env explicitly so DATABASE_URL is available
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found or error loading it:", err)
	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL is empty")
	}

	conn, err := pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer conn.Close(context.Background())

	// Example query to test connection
	var version string
	if err := conn.QueryRow(context.Background(), "SELECT version()").Scan(&version); err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	log.Println("Connected to:", version)
}
