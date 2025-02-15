package db

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	var err error
	DB, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	fmt.Println("Database connected successfully!")
}
