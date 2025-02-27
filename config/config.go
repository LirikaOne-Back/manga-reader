package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerAddress string
	DBType        string
	DBSource      string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	JWTSecret     string
}

func LoadConfig() Config {
	_ = godotenv.Load()

	return Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
		DBType:        getEnv("DB_TYPE", "sqlite"),
		DBSource:      getEnv("DB_SOURCE", "manga.db"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 1),
		JWTSecret:     getEnv("JWT_SECRET", "secret"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(strings.TrimSpace(value)) == 0 {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if val, err := strconv.Atoi(valueStr); err == nil {
		return val
	}
	return defaultValue

}
