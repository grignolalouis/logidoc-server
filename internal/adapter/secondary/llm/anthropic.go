package llm

import (
	"trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/model/anthropic"

	"github.com/logidoc/logidoc-server/internal/config"
)

func newAnthropic(cfg config.LLMConfig) model.Model {
	return anthropic.New(cfg.Model, anthropic.WithAPIKey(cfg.APIKey))
}
