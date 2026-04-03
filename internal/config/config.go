package config

import "time"

type Config struct {
	App     AppConfig
	HTTP    HTTPConfig
	MCP     MCPConfig
	Mongo   MongoConfig
	LLM     LLMConfig
	Indexer IndexerConfig
	Logger  LoggerConfig
}

type AppConfig struct {
	Version string
	APIKey  string // optional, enables API key auth if set
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
	Provider string // anthropic, openai, mistral, xai, groq, ollama
	Model    string
	APIKey   string
	BaseURL  string // optional override for OpenAI-compatible providers
}

type IndexerConfig struct {
	MaxPagesPerNode  int
	MaxTokensPerNode int
	TOCCheckPages    int
}

type LoggerConfig struct {
	Level  string
	Format string
}
