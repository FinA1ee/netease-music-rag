package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                    string
	PostgresDSN             string
	GeminiAPIKey            string
	LLMModel                string
	EmbeddingModel          string
	NeteaseAPIURL           string
	NeteasePhone            string
	SearchMinPop            float32
	SearchDistanceThreshold float32
	SearchVectorWeight      float32
	SearchLexicalWeight     float32
}

func Load() *Config {
	// Load .env file if present (no-op when env vars are already set, e.g. in Docker)
	if err := godotenv.Load(); err != nil {
		// Not an error — .env is optional in containerised deployments
	}

	dsn := getEnv("POSTGRES_DSN", "")
	if dsn == "" {
		sslmode := getEnv("POSTGRES_SSLMODE", "disable")
		// Build standard GORM DSN string to handle passwords with symbols seamlessly.
		// Used when POSTGRES_DSN is not explicitly provided.
		dsn = "postgresql://" +
			getEnv("POSTGRES_USER", "postgres") + ":" +
			getEnv("POSTGRES_PASSWORD", "postgres") + "@" +
			getEnv("POSTGRES_HOST", "postgres") + ":" +
			getEnv("POSTGRES_PORT", "5432") + "/" +
			getEnv("POSTGRES_DB", "postgres") +
			"?sslmode=" + sslmode
	}

	return &Config{
		Port:           getEnv("PORT", "8080"),
		PostgresDSN:    dsn,
		GeminiAPIKey:   getEnv("GEMINI_API_KEY", ""),
		LLMModel:       getEnv("LLM_MODEL", ""),
		EmbeddingModel: getEnv("EMBEDDING_MODEL", ""),
		// Default to the Docker Compose service name; override to http://localhost:3000 locally
		NeteaseAPIURL:           getEnv("NETEASE_API_URL", "http://netease-api:3000"),
		NeteasePhone:            getEnv("NETEASE_PHONE", ""),
		SearchMinPop:            getEnvFloat32("SEARCH_MIN_POPULARITY", 30),
		SearchDistanceThreshold: getEnvFloat32("SEARCH_DISTANCE_THRESHOLD", 1.2),
		SearchVectorWeight:      getEnvFloat32("SEARCH_VECTOR_WEIGHT", 1.0),
		SearchLexicalWeight:     getEnvFloat32("SEARCH_LEXICAL_WEIGHT", 0.2),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func getEnvFloat32(key string, fallback float32) float32 {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return fallback
	}
	return float32(parsed)
}
