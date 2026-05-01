package config

import (
	"os"
	"strconv"
)

const defaultDBPort = 3311

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
		Host:     os.Getenv("DB_HOST"),
		Port:     dbPort(),
		User:     os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}
}

func dbPort() int {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil || port <= 0 {
		return defaultDBPort
	}

	return port
}
