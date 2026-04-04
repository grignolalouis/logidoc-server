package config

import (
	"fmt"
	"strings"
	"time"
)

type Config struct {
	App     AppConfig
	HTTP    HTTPConfig
	MCP     MCPConfig
	Mongo   MongoConfig
	LLM     LLMConfig
	Vision  VisionConfig
	Indexer IndexerConfig
	Logger  LoggerConfig
}

func (c *Config) validate() error {
	if c.LLM.APIKey == "" {
		return fmt.Errorf("LLM_API_KEY is required")
	}
	if c.LLM.Provider == "" {
		return fmt.Errorf("LLM_PROVIDER is required")
	}
	if c.LLM.Model == "" {
		return fmt.Errorf("LLM_MODEL is required")
	}
	if !strings.HasPrefix(c.Mongo.URI, "mongodb") {
		return fmt.Errorf("MONGO_URI must start with mongodb:// or mongodb+srv://")
	}
	if c.HTTP.BodyLimitMB <= 0 {
		return fmt.Errorf("HTTP_BODY_LIMIT_MB must be positive")
	}
	if c.HTTP.RateLimit <= 0 {
		return fmt.Errorf("HTTP_RATE_LIMIT must be positive")
	}
	return nil
}

type AppConfig struct {
	Version string
	APIKey  string
}

type HTTPConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	CORSOrigins  string
	RateLimit    int
	BodyLimitMB  int
}

type MCPConfig struct {
	Addr      string
	Transport string
	Stdio     bool
}

type MongoConfig struct {
	URI      string
	Database string
}

type LLMConfig struct {
	Provider string
	Model    string
	APIKey   string
	BaseURL  string
}

type VisionConfig struct {
	Provider string
	Model    string
	APIKey   string
	BaseURL  string
	Enabled  bool // true if VISION_PROVIDER is set
}

type IndexerConfig struct {
	MaxPagesPerNode        int
	MaxTokensPerNode       int
	TOCCheckPages          int
	EnableTableExtraction  bool
	EnableImageDescription bool
}

type LoggerConfig struct {
	Level  string
	Format string
}
