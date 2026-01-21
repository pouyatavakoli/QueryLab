package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost string
	DBPort string

	DBAdminUser     string
	DBAdminPassword string

	DBSandboxUser     string
	DBSandboxPassword string

	DBName     string
	ServerPort string
	InitSQL    string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "5432"),

		DBAdminUser:     getEnv("DB_ADMIN_USER", ""),
		DBAdminPassword: getEnv("DB_ADMIN_PASSWORD", ""),

		DBSandboxUser:     getEnv("DB_SANDBOX_USER", ""),
		DBSandboxPassword: getEnv("DB_SANDBOX_PASSWORD", ""),

		DBName:     getEnv("DB_NAME", "querylab"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		InitSQL:    getEnv("INIT_SQL", "init.sql"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
