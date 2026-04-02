// Package llm provides LLM model creation from configuration.
//
// Supported providers:
//
//	anthropic  → Anthropic API (Claude models)
//	openai     → OpenAI API (GPT models)
//	mistral    → Mistral AI (OpenAI-compatible)
//	xai        → xAI / Grok (OpenAI-compatible)
//	groq       → Groq (OpenAI-compatible)
//	ollama     → Local Ollama (OpenAI-compatible)
//	custom     → Any OpenAI-compatible API (requires LLM_BASE_URL)
//
// Most providers use the OpenAI-compatible protocol with a different base URL.
// Only Anthropic requires a dedicated SDK.
package llm

import (
	"fmt"
	"log/slog"

	"trpc.group/trpc-go/trpc-agent-go/model"

	"github.com/logidoc/logidoc-server/internal/config"
)

// Known base URLs for OpenAI-compatible providers.
var knownBaseURLs = map[string]string{
	"mistral": "https://api.mistral.ai/v1",
	"xai":     "https://api.x.ai/v1",
	"groq":    "https://api.groq.com/openai/v1",
	"ollama":  "http://host.docker.internal:11434/v1",
}

// NewModel creates a model.Model from config.
//
// Provider resolution:
//  1. "anthropic" → Anthropic SDK
//  2. "openai" → OpenAI SDK (optionally with custom base URL)
//  3. Known name (mistral, xai, groq, ollama) → OpenAI SDK + known base URL
//  4. Unknown name → OpenAI SDK + LLM_BASE_URL (required)
func NewModel(cfg config.LLMConfig, logger *slog.Logger) (model.Model, error) {
	logger.Info("initializing LLM", "provider", cfg.Provider, "model", cfg.Model)

	switch cfg.Provider {
	case "anthropic":
		return newAnthropic(cfg), nil

	case "openai":
		if cfg.BaseURL != "" {
			return newOpenAICompat(cfg, cfg.BaseURL), nil
		}
		return newOpenAI(cfg), nil

	default:
		// Known OpenAI-compatible provider
		if baseURL, ok := knownBaseURLs[cfg.Provider]; ok {
			if cfg.BaseURL != "" {
				baseURL = cfg.BaseURL // user override
			}
			return newOpenAICompat(cfg, baseURL), nil
		}

		// Unknown — needs explicit base URL
		if cfg.BaseURL != "" {
			return newOpenAICompat(cfg, cfg.BaseURL), nil
		}

		return nil, fmt.Errorf(
			"unknown provider %q — use one of: openai, anthropic, mistral, xai, groq, ollama, or set LLM_BASE_URL for a custom OpenAI-compatible endpoint",
			cfg.Provider,
		)
	}
}
