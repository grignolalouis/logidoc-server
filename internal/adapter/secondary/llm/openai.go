package llm

import (
	"trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/model/openai"

	"github.com/logidoc/logidoc-server/internal/config"
)

func newOpenAI(cfg config.LLMConfig) model.Model {
	return openai.New(cfg.Model, openai.WithAPIKey(cfg.APIKey))
}

// newOpenAICompat creates an OpenAI-compatible model with a custom base URL.
// Works for Mistral, Grok, Groq, Ollama, and any OpenAI-compatible API.
func newOpenAICompat(cfg config.LLMConfig, baseURL string) model.Model {
	return openai.New(cfg.Model,
		openai.WithAPIKey(cfg.APIKey),
		openai.WithBaseURL(baseURL),
	)
}
