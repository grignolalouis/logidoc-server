package indexer

import (
	"time"

	"trpc.group/trpc-go/trpc-agent-go/event"
	"trpc.group/trpc-go/trpc-agent-go/model"
)

// Metrics captures indexation performance data.
type Metrics struct {
	Duration         time.Duration
	LLMCalls         int
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	PagesRead        int
	PagesTotal       int
	SectionsFound    int
}

// AddUsage accumulates token usage from an LLM response.
func (m *Metrics) AddUsage(usage *model.Usage) {
	if usage == nil {
		return
	}
	m.PromptTokens += usage.PromptTokens
	m.CompletionTokens += usage.CompletionTokens
	m.TotalTokens += usage.TotalTokens
	m.LLMCalls++
}

// CollectEvents drains the event channel, accumulates metrics, and returns
// the concatenated text content from all events.
func CollectEvents(eventChan <-chan *event.Event, m *Metrics) (string, error) {
	var buf []byte

	for ev := range eventChan {
		if ev.Error != nil {
			return "", ev.Error
		}
		if ev.Response == nil || len(ev.Response.Choices) == 0 {
			continue
		}
		m.AddUsage(ev.Response.Usage)

		c := ev.Response.Choices[0]
		if c.Delta.Content != "" {
			buf = append(buf, c.Delta.Content...)
		}
		if c.Message.Content != "" {
			buf = append(buf, c.Message.Content...)
		}
	}

	return string(buf), nil
}
