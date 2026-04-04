package config

import (
	"os"
)

type Config struct {
	Port           string
	PostgresDSN    string
	GeminiAPIKey   string
	LLMModel       string
	EmbeddingModel string
	NeteaseAPIURL  string
	NeteasePhone   string
	NeteasePass    string
	NeteaseAppId   string
	NeteaseAppKey  string
}

func Load() *Config {
	// Build standard GORM DSN string to handle passwords with symbols seamlessly

	dsn := "postgresql://" +
		getEnv("POSTGRES_USER", "postgres") + ":" +
		getEnv("POSTGRES_PASSWORD", "otakufrosT1997") + "@" +
		getEnv("POSTGRES_HOST", "db.zqvjfjrhloemqbdtsqwd.supabase.co") + ":" +
		getEnv("POSTGRES_PORT", "5432") + "/" +
		getEnv("POSTGRES_DB", "postgres") +
		"?sslmode=disable"

	return &Config{
		Port:           getEnv("PORT", "8080"),
		PostgresDSN:    dsn,
		GeminiAPIKey:   getEnv("GEMINI_API_KEY", ""),
		LLMModel:       getEnv("LLM_MODEL", "gemini-2.5-flash"),
		EmbeddingModel: getEnv("EMBEDDING_MODEL", "gemini-embedding-001"),
		NeteaseAPIURL:  getEnv("NETEASE_API_URL", "http://localhost:3000"),
		NeteasePhone:   getEnv("NETEASE_PHONE", ""),
		NeteasePass:    getEnv("NETEASE_PASSWORD", ""),
		NeteaseAppId:   getEnv("NETEASE_APP_ID", ""),
		NeteaseAppKey:  getEnv("NETEASE_APP_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
