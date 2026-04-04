package index

import (
	"time"

	"trpc.group/trpc-go/trpc-agent-go/event"
	"trpc.group/trpc-go/trpc-agent-go/model"
)

type Metrics struct {
	Duration         time.Duration
	PagesTotal       int
	PagesRead        int
	SectionsFound    int

	// Agent (structure detection, TOC, chunking, subdivision)
	AgentCalls           int
	AgentPromptTokens    int
	AgentCompletionTokens int
	AgentTotalTokens     int

	// Vision (table extraction, image description)
	VisionCalls           int
	VisionPromptTokens    int
	VisionCompletionTokens int
	VisionTotalTokens     int
}

func (m *Metrics) TotalCalls() int  { return m.AgentCalls + m.VisionCalls }
func (m *Metrics) TotalTokens() int { return m.AgentTotalTokens + m.VisionTotalTokens }

func (m *Metrics) addAgent(usage *model.Usage) {
	if usage == nil {
		return
	}
	m.AgentPromptTokens += usage.PromptTokens
	m.AgentCompletionTokens += usage.CompletionTokens
	m.AgentTotalTokens += usage.TotalTokens
	m.AgentCalls++
}

func (m *Metrics) AddVision(usage *model.Usage) {
	if usage == nil {
		return
	}
	m.VisionPromptTokens += usage.PromptTokens
	m.VisionCompletionTokens += usage.CompletionTokens
	m.VisionTotalTokens += usage.TotalTokens
	m.VisionCalls++
}

// AddUsage adds to agent metrics (backward compat for chunker/subdivide).
func (m *Metrics) AddUsage(usage *model.Usage) {
	m.addAgent(usage)
}

func (m *Metrics) merge(other *Metrics) {
	if other == nil {
		return
	}
	m.AgentCalls += other.AgentCalls
	m.AgentPromptTokens += other.AgentPromptTokens
	m.AgentCompletionTokens += other.AgentCompletionTokens
	m.AgentTotalTokens += other.AgentTotalTokens
	m.VisionCalls += other.VisionCalls
	m.VisionPromptTokens += other.VisionPromptTokens
	m.VisionCompletionTokens += other.VisionCompletionTokens
	m.VisionTotalTokens += other.VisionTotalTokens
	m.PagesRead += other.PagesRead
}

func CollectEvents(eventChan <-chan *event.Event, m *Metrics) (string, error) {
	var buf []byte

	for ev := range eventChan {
		if ev.Error != nil {
			return "", ev.Error
		}
		if ev.Response == nil || len(ev.Response.Choices) == 0 {
			continue
		}
		m.addAgent(ev.Response.Usage)

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
