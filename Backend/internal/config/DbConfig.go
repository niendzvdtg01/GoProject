package config

import (
	"os"
)

// config for database connection
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func NewDBConfig() *DBConfig {
	return &DBConfig{}
}

func (c *DBConfig) GetDSN() DBConfig {
	return DBConfig{
		Host:     os.Getenv("HOST"),
		Port:     3311,
		User:     os.Getenv("USERNAME"),
		Password: os.Getenv("PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}
}
