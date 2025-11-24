package database

import (
	"log"
	"os"
	"fmt"
	"database/sql"
	_ "github.com/lib/pq" 
)

var PostgresDB *sql.DB
func ConnectPostgres() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	PostgresDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	fmt.Println("Connected to PostgreSQL")
	fmt.Println("DB Postgresql :", os.Getenv("DB_NAME"))
}