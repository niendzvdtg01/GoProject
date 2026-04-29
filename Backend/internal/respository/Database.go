package database

import (
	"backend/internal/config"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// Conect to database
func ConnectDB(cfg config.DBConfig) error {
	var err error

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	fmt.Printf("Connecting to database with DSN: %s\n", dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %v", err)
	}

	// Connection pool setting
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	DB = db

	return nil
}

// Close the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
