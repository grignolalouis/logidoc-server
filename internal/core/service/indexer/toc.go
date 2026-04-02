package indexer

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	"trpc.group/trpc-go/trpc-agent-go/agent/llmagent"
	"trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/runner"
	"trpc.group/trpc-go/trpc-agent-go/session/inmemory"
	"trpc.group/trpc-go/trpc-agent-go/tool"
	"trpc.group/trpc-go/trpc-agent-go/tool/function"

	"github.com/logidoc/logidoc-server/pkg/jsonutil"
	"github.com/logidoc/logidoc-server/pkg/ptr"
)

// TOCResult holds the result of TOC detection.
type TOCResult struct {
	Sections []FlatSection
	Metrics  Metrics
	Found    bool
}

// DetectTOC uses an LLM agent to read the first pages and extract the TOC.
// Returns Found=false if no TOC is detected (caller should fall back to chunking).
func DetectTOC(ctx context.Context, llm model.Model, pages *Pages) (*TOCResult, error) {
	var pagesRead atomic.Int32
	tools := pageTools(pages, &pagesRead)

	ag := llmagent.New("toc-detector",
		llmagent.WithModel(llm),
		llmagent.WithInstruction(TOCAgentInstruction),
		llmagent.WithTools(tools),
		llmagent.WithGenerationConfig(model.GenerationConfig{
			Temperature: ptr.Float64(0.1),
			MaxTokens:   ptr.Int(8000),
			Stream:      false,
		}),
		llmagent.WithMaxLLMCalls(10),
	)

	r := runner.NewRunner("logidoc-toc", ag, runner.WithSessionService(inmemory.NewSessionService()))
	defer r.Close()

	msg := model.NewUserMessage(fmt.Sprintf(
		"Detect the structure of: %q (%d pages).", pages.Filename, pages.Count(),
	))

	eventChan, err := r.Run(ctx, "indexer", "toc-detect", msg)
	if err != nil {
		return nil, fmt.Errorf("run toc agent: %w", err)
	}

	var m Metrics
	result, evErr := CollectEvents(eventChan, &m)
	m.PagesRead = int(pagesRead.Load())

	if evErr != nil {
		return &TOCResult{Metrics: m}, fmt.Errorf("toc agent: %w", evErr)
	}

	// Check for NO_TOC signal
	if result == "" || strings.Contains(result, "NO_TOC") {
		return &TOCResult{Metrics: m, Found: false}, nil
	}

	sections, err := jsonutil.ParseArray[FlatSection](result)
	if err != nil {
		// Not parseable — treat as no TOC, caller will fall back to chunking
		return &TOCResult{Metrics: m, Found: false}, nil
	}

	return &TOCResult{Sections: sections, Metrics: m, Found: true}, nil
}

// ---- Page tools for the agent ----

type readPagesIn struct {
	Start int `json:"start"`
	End   int `json:"end"`
}
type readPagesOut struct{ Text string `json:"text"` }
type pageCountOut struct{ Total int `json:"total"` }

func pageTools(pages *Pages, pagesRead *atomic.Int32) []tool.Tool {
	return []tool.Tool{
		function.NewFunctionTool(
			func(_ context.Context, _ struct{}) (pageCountOut, error) {
				return pageCountOut{Total: pages.Count()}, nil
			},
			function.WithName("get_page_count"),
			function.WithDescription("Returns the total number of pages"),
		),
		function.NewFunctionTool(
			func(_ context.Context, in readPagesIn) (readPagesOut, error) {
				if in.Start < 1 || in.End < in.Start {
					return readPagesOut{}, fmt.Errorf("invalid range: %d-%d", in.Start, in.End)
				}
				if in.End-in.Start+1 > 10 {
					return readPagesOut{}, fmt.Errorf("max 10 pages per call")
				}
				pagesRead.Add(int32(in.End - in.Start + 1))
				return readPagesOut{Text: pages.Read(in.Start, in.End)}, nil
			},
			function.WithName("read_pages"),
			function.WithDescription("Read text content of pages (1-indexed, inclusive, max 10)"),
		),
	}
}
