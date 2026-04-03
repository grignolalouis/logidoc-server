package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"trpc.group/trpc-go/trpc-agent-go/model"

	"github.com/logidoc/logidoc-server/pkg/jsonutil"
	"github.com/logidoc/logidoc-server/pkg/ptr"
)

const defaultChunkSize = 10 // pages per chunk

// ProcessNoTOC scans the entire document in sequential chunks,
// asking the LLM to extract structure from each chunk.
// Each chunk sees the accumulated sections from previous chunks for continuity.
func ProcessNoTOC(ctx context.Context, llm model.Model, pages *Pages, logger *slog.Logger) ([]FlatSection, *Metrics, error) {
	m := &Metrics{}
	var allSections []FlatSection
	total := pages.Count()

	for start := 1; start <= total; start += defaultChunkSize {
		end := start + defaultChunkSize - 1
		if end > total {
			end = total
		}

		text := pages.Read(start, end)
		m.PagesRead += end - start + 1

		prompt := buildChunkPrompt(allSections, start, end, total, text)

		newSections, usage, err := callLLM(ctx, llm, prompt)
		if err != nil {
			logger.Warn("chunk error, skipping", "pages", fmt.Sprintf("%d-%d", start, end), "error", err)
			continue
		}

		m.AddUsage(usage)
		allSections = append(allSections, newSections...)

		logger.Debug("chunk processed",
			"pages", fmt.Sprintf("%d-%d", start, end),
			"new_sections", len(newSections),
			"total_sections", len(allSections),
		)
	}

	if len(allSections) == 0 {
		return nil, m, fmt.Errorf("no sections detected in document")
	}

	return allSections, m, nil
}

func buildChunkPrompt(existing []FlatSection, start, end, total int, text string) string {
	if len(existing) == 0 {
		return fmt.Sprintf(ChunkInitPrompt, start, end, total, text)
	}
	prev, _ := json.Marshal(existing)
	return fmt.Sprintf(ChunkContinuePrompt, string(prev), start, end, total, text)
}

const maxRetries = 3

func callLLM(ctx context.Context, llm model.Model, prompt string) ([]FlatSection, *model.Usage, error) {
	req := &model.Request{
		Messages: []model.Message{
			model.NewUserMessage(prompt),
		},
		GenerationConfig: model.GenerationConfig{
			Temperature: ptr.Float64(0.1),
			MaxTokens:   ptr.Int(4000),
			Stream:      false,
		},
	}

	var lastErr error
	for attempt := range maxRetries {
		sections, usage, err := doLLMCall(ctx, llm, req)
		if err == nil {
			return sections, usage, nil
		}
		lastErr = err
		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
		}
	}
	return nil, nil, fmt.Errorf("llm call failed after %d retries: %w", maxRetries, lastErr)
}

func doLLMCall(ctx context.Context, llm model.Model, req *model.Request) ([]FlatSection, *model.Usage, error) {
	respChan, err := llm.GenerateContent(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("llm call: %w", err)
	}

	var content string
	var usage *model.Usage
	for resp := range respChan {
		if resp.Error != nil {
			return nil, nil, fmt.Errorf("llm error: %s", resp.Error.Message)
		}
		usage = resp.Usage
		if len(resp.Choices) > 0 {
			content += resp.Choices[0].Message.Content
			content += resp.Choices[0].Delta.Content
		}
	}

	sections, err := jsonutil.ParseArray[FlatSection](content)
	return sections, usage, err
}
