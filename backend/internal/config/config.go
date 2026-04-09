package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	PostgresDSN    string
	GeminiAPIKey   string
	LLMModel       string
	EmbeddingModel string
	NeteaseAPIURL  string
	NeteasePhone   string
	NeteaseAppId   string
	NeteaseAppKey  string
}

func Load() *Config {
	// Load .env file if present (no-op when env vars are already set, e.g. in Docker)
	if err := godotenv.Load(); err != nil {
		// Not an error — .env is optional in containerised deployments
	}

	// Build standard GORM DSN string to handle passwords with symbols seamlessly
	dsn := "postgresql://" +
		getEnv("POSTGRES_USER", "postgres") + ":" +
		getEnv("POSTGRES_PASSWORD", "otakufrosT1997") + "@" +
		getEnv("POSTGRES_HOST", "db.zqvjfjrhloemqbdtsqwd.supabase.co") + ":" +
		getEnv("POSTGRES_PORT", "5432") + "/" +
		getEnv("POSTGRES_DB", "postgres") +
		"?sslmode=require"

	return &Config{
		Port:           getEnv("PORT", "8080"),
		PostgresDSN:    dsn,
		GeminiAPIKey:   getEnv("GEMINI_API_KEY", ""),
		LLMModel:       getEnv("LLM_MODEL", ""),
		EmbeddingModel: getEnv("EMBEDDING_MODEL", ""),
		NeteaseAPIURL:  getEnv("NETEASE_API_URL", ""),
		NeteasePhone:   getEnv("NETEASE_PHONE", ""),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
