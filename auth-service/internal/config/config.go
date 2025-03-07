package config

import (
	"auth-service/internal/logger"
	"github.com/joho/godotenv"
	"os"
)

func LoadConfig(logs *logger.Logger) map[string]string {
	if err := godotenv.Load("../.env"); err != nil {
		logs.Error.Println("No .env file found, using system environment variables")
	}
	return map[string]string{
		"POSTGRES_HOST":     os.Getenv("POSTGRES_HOST"),
		"POSTGRES_PORT":     os.Getenv("POSTGRES_PORT"),
		"POSTGRES_USER":     os.Getenv("POSTGRES_USER"),
		"POSTGRES_PASSWORD": os.Getenv("POSTGRES_PASSWORD"),
		"POSTGRES_DB":       os.Getenv("POSTGRES_DB"),
		"JWT_SECRET":        os.Getenv("JWT_SECRET"),
	}
}
