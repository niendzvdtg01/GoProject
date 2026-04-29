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
	// port, err := strconv.Atoi(os.Getenv("DB_USERNAME"))

	// if err != nil {
	// 	fmt.Println("Incorrect port", err)
	// }
	return DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     3311,
		User:     os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}
}
