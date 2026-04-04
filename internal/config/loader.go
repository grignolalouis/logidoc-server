package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func Load() (*Config, error) {
	godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Version: getEnv("APP_VERSION", "dev"),
			APIKey:  getEnv("API_KEY", ""),
		},
		HTTP: HTTPConfig{
			Addr:         getEnv("HTTP_ADDR", ":7042"),
			ReadTimeout:  getDuration("HTTP_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
			CORSOrigins:  getEnv("HTTP_CORS_ORIGINS", "*"),
			RateLimit:    getInt("HTTP_RATE_LIMIT", 100),
			BodyLimitMB:  getInt("HTTP_BODY_LIMIT_MB", 50),
		},
		MCP: MCPConfig{
			Addr:      getEnv("MCP_ADDR", ":7043"),
			Transport: getEnv("MCP_TRANSPORT", "streamable_http"),
			Stdio:     getBool("MCP_STDIO", false),
		},
		Mongo: MongoConfig{
			URI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGO_DATABASE", "logidoc"),
		},
		LLM: LLMConfig{
			Provider: getEnv("LLM_PROVIDER", "openai"),
			Model:    getEnv("LLM_MODEL", "gpt-4o"),
			APIKey:   getEnv("LLM_API_KEY", ""),
			BaseURL:  getEnv("LLM_BASE_URL", ""),
		},
		Vision: buildVisionConfig(),
		Indexer: IndexerConfig{
			MaxPagesPerNode:        getInt("INDEXER_MAX_PAGES_PER_NODE", 20),
			MaxTokensPerNode:       getInt("INDEXER_MAX_TOKENS_PER_NODE", 20000),
			TOCCheckPages:          getInt("INDEXER_TOC_CHECK_PAGES", 20),
			EnableTableExtraction:  getBool("INDEXER_ENABLE_TABLE_EXTRACTION", false),
			EnableImageDescription: getBool("INDEXER_ENABLE_IMAGE_DESCRIPTION", false),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	return cfg, cfg.validate()
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func buildVisionConfig() VisionConfig {
	provider := getEnv("VISION_PROVIDER", "")
	if provider == "" {
		return VisionConfig{}
	}
	return VisionConfig{
		Provider: provider,
		Model:    getEnv("VISION_MODEL", ""),
		APIKey:   getEnv("VISION_API_KEY", getEnv("LLM_API_KEY", "")),
		BaseURL:  getEnv("VISION_BASE_URL", ""),
		Enabled:  true,
	}
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
