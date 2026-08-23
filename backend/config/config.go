package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	AppPort     string
	FrontendURL string

	DBDriver string
	DBHost   string
	DBPort   string
	DBUser   string
	DBPass   string
	DBName   string

	RedisHost string
	RedisPort string
	RedisPass string
	RedisDB   int

	SessionMaxAge        int
	SessionMaxConcurrent int

	StorageMaxSize int64

	S3Endpoint  string
	S3Region    string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string

	LogCleanupDays           int
	RateLimitLogin           int
	RateLimitLoginLockoutMin int
	RateLimitAPI             int

	AnthropicAPIKey string
	AnthropicModel  string
}

var Cfg *Config

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on environment variables")
	}

	Cfg = &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		AppPort:     getEnv("APP_PORT", "8500"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:8501,http://localhost:8502"),

		DBDriver: getEnv("DB_DRIVER", "postgres"),
		DBHost:   getEnv("DB_HOST", "localhost"),
		DBPort:   getEnv("DB_PORT", "5433"),
		DBUser:   getEnv("DB_USER", "postgres"),
		DBPass:   getEnv("DB_PASS", "password"),
		DBName:   getEnv("DB_NAME", "baseadmin"),

		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),
		RedisPass: getEnv("REDIS_PASS", ""),
		RedisDB:   getEnvInt("REDIS_DB", 0),

		SessionMaxAge:        getEnvInt("SESSION_MAX_AGE", 86400),
		SessionMaxConcurrent: getEnvInt("SESSION_MAX_CONCURRENT", 3),

		StorageMaxSize: int64(getEnvInt("STORAGE_MAX_SIZE", 2097152)),

		S3Endpoint:  getEnv("S3_ENDPOINT", ""),
		S3Region:    getEnv("S3_REGION", ""),
		S3AccessKey: getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey: getEnv("S3_SECRET_KEY", ""),
		S3Bucket:    getEnv("S3_BUCKET", ""),

		LogCleanupDays:           getEnvInt("LOG_CLEANUP_DAYS", 90),
		RateLimitLogin:           getEnvInt("RATE_LIMIT_LOGIN", 5),
		RateLimitLoginLockoutMin: getEnvInt("RATE_LIMIT_LOGIN_LOCKOUT_MINUTES", 15),
		RateLimitAPI:             getEnvInt("RATE_LIMIT_API", 100),

		AnthropicAPIKey: getEnv("ANTHROPIC_API_KEY", ""),
		AnthropicModel:  getEnv("ANTHROPIC_MODEL", "claude-opus-5"),
	}

	return Cfg
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
