package database

import (
	"backend/internal/config"
	"database/sql"
	"fmt"
)

var DB *sql.DB

// Conect to database
func ConnectDB(cfg config.DBConfig) error {
	var err error

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	DB = db

	return nil
}
