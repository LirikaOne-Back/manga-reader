package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerAddress string
	DBType        string
	DBSource      string
	PgHost        string
	PgPort        int
	PgUser        string
	PgPassword    string
	PgDBName      string
	PgSSLMode     string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	JWTSecret     string
}

func LoadConfig() Config {
	_ = godotenv.Load()

	inDocker := false
	if _, err := os.Stat("/.dockerenv"); err == nil {
		inDocker = true
	}

	pgHost := getEnv("PG_HOST", "localhost")

	if inDocker && pgHost == "localhost" {
		pgHost = "postgres"
	}

	return Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
		DBType:        getEnv("DB_TYPE", "sqlite"),
		DBSource:      getEnv("DB_SOURCE", "manga.db"),
		PgHost:        pgHost,
		PgPort:        getEnvAsInt("PG_PORT", 5432),
		PgUser:        getEnv("PG_USER", "postgres"),
		PgPassword:    getEnv("PG_PASSWORD", ""),
		PgDBName:      getEnv("PG_DBNAME", "manga_reader"),
		PgSSLMode:     getEnv("PG_SSLMODE", "disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 1),
		JWTSecret:     getEnv("JWT_SECRET", "secret"),
	}
}

func (c *Config) PostgresConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.PgHost, c.PgPort, c.PgUser, c.PgPassword, c.PgDBName, c.PgSSLMode)
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

func (c *Config) PostgresMigrationURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.PgUser, c.PgPassword, c.PgHost, c.PgPort, c.PgDBName, c.PgSSLMode)
}
