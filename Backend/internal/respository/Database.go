package database

import (
	"backend/internal/config"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	maxOpenConnections = 25
	maxIdleConnections = 5
	connectionLifetime = 5 * time.Minute
)

var DB *sql.DB

func ConnectDB(cfg config.DBConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open database connection: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConnections)
	db.SetMaxIdleConns(maxIdleConnections)
	db.SetConnMaxLifetime(connectionLifetime)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	DB = db
	return nil
}

func CloseDB() error {
	if DB == nil {
		return nil
	}

	return DB.Close()
}
