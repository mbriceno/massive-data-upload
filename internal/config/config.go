package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDSN      string
	NumWorkers int
}

// Load lee el entorno
func Load() *Config {
	// Intentar cargar el archivo .env si existe (en prod puede no existir y usarse env vars reales)
	_ = godotenv.Load()

	host := getEnv("DB_HOST", "localhost")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "secret")
	dbname := getEnv("DB_NAME", "mi_empresa")
	port := getEnv("DB_PORT", "5432")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	workersStr := getEnv("NUM_WORKERS", "4")
	workers, err := strconv.Atoi(workersStr)
	if err != nil {
		workers = 4 // Fallback en caso de que pongan texto inválido en el .env
	}

	return &Config{
		DBDSN:      dsn,
		NumWorkers: workers,
	}
}

// Función auxiliar para manejar valores por defecto
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
