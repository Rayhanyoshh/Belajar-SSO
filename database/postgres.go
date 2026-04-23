package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	// Membaca pengaturan dari Docker (Environment Variables)
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "personal"
	}
	
	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		dbPass = "crudence"
	}

	connStr := fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=go_sso sslmode=disable", dbHost, dbUser, dbPass)
	
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Gagal membuka koneksi database: ", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Database tidak merespons (Ping gagal): ", err)
	}

	fmt.Println("✅ [SSO] Terhubung ke PostgreSQL (Database: go_sso)!")
	
	createTables()
}

func createTables() {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL
	);`
	
	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal("Gagal membuat tabel users: ", err)
	}
	fmt.Println("✅ [SSO] Tabel database siap digunakan.")
}
