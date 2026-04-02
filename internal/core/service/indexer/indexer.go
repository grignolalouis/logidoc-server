// Package indexer implements the document indexation pipeline.
//
// Two paths depending on document structure:
//   - TOC detected: Agent reads first pages, extracts structure directly (fast, cheap)
//   - No TOC: Sequential chunking scans every page in groups of 10 (thorough, costly)
//
// In both cases, the LLM produces a flat list of sections with page ranges.
// Go code then builds the tree and fills text from parsed pages (zero LLM cost).
package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"trpc.group/trpc-go/trpc-agent-go/model"

	"github.com/logidoc/logidoc-server/internal/core/domain"
	"github.com/logidoc/logidoc-server/internal/core/port"
)

// Service orchestrates document indexation.
type Service struct {
	llm     model.Model
	docRepo port.DocumentRepository
	idxRepo port.IndexRepository
	logger  *slog.Logger
}

// NewService creates a new indexer service.
func NewService(llm model.Model, docRepo port.DocumentRepository, idxRepo port.IndexRepository, logger *slog.Logger) *Service {
	return &Service{llm: llm, docRepo: docRepo, idxRepo: idxRepo, logger: logger}
}

// Index parses a file, detects structure, fills text, and saves the index.
func (s *Service) Index(ctx context.Context, docID string, filename string, data []byte) error {
	start := time.Now()
	s.logger.Info("starting indexation", "doc_id", docID, "filename", filename, "size", len(data))

	if err := s.docRepo.UpdateStatus(ctx, docID, domain.StatusIndexing, ""); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	// Step 1: Parse file into pages (Go, no LLM)
	pages, err := ParseFile(filename, data)
	if err != nil {
		s.fail(ctx, docID, err)
		return fmt.Errorf("parse file: %w", err)
	}
	s.logger.Info("parsed document", "doc_id", docID, "pages", pages.Count())

	// Step 2: Detect structure (TOC agent or sequential chunking)
	sections, metrics, err := s.detectStructure(ctx, pages)
	if err != nil {
		s.fail(ctx, docID, err)
		return err
	}

	// Step 3: Calibrate page numbers (logical → physical) then build tree
	before := sections[0].StartPage
	sections = CalibratePages(sections, pages)
	if sections[0].StartPage != before {
		s.logger.Info("page offset detected",
			"doc_id", docID,
			"offset", sections[0].StartPage-before,
			"before", before,
			"after", sections[0].StartPage,
		)
	}
	tree := BuildTree(sections, pages.Count())
	nodes := FillText(tree, pages)

	// Step 4: Save
	if err := s.idxRepo.Save(ctx, &domain.Index{DocID: docID, Tree: nodes, Version: 1}); err != nil {
		s.fail(ctx, docID, err)
		return fmt.Errorf("save index: %w", err)
	}
	if err := s.docRepo.UpdateStatus(ctx, docID, domain.StatusReady, ""); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	// Log metrics
	metrics.Duration = time.Since(start)
	metrics.PagesTotal = pages.Count()
	metrics.SectionsFound = CountNodes(nodes)

	s.logger.Info("indexation complete",
		"doc_id", docID,
		"duration", metrics.Duration.Round(time.Millisecond),
		"pages", metrics.PagesTotal,
		"sections", metrics.SectionsFound,
		"llm_calls", metrics.LLMCalls,
		"prompt_tokens", metrics.PromptTokens,
		"completion_tokens", metrics.CompletionTokens,
		"total_tokens", metrics.TotalTokens,
		"pages_read", metrics.PagesRead,
	)
	return nil
}

// detectStructure tries the TOC agent first, falls back to sequential chunking.
func (s *Service) detectStructure(ctx context.Context, pages *Pages) ([]FlatSection, *Metrics, error) {
	// Try TOC detection
	tocResult, err := DetectTOC(ctx, s.llm, pages)
	if err != nil {
		return nil, nil, err
	}

	if tocResult.Found {
		s.logger.Info("TOC detected", "sections", len(tocResult.Sections))
		return tocResult.Sections, &tocResult.Metrics, nil
	}

	// No TOC → sequential chunking
	s.logger.Info("no TOC detected, switching to sequential chunking")
	sections, chunkMetrics, err := ProcessNoTOC(ctx, s.llm, pages, s.logger)
	if err != nil {
		return nil, nil, err
	}

	// Merge metrics from TOC attempt + chunking
	chunkMetrics.PromptTokens += tocResult.Metrics.PromptTokens
	chunkMetrics.CompletionTokens += tocResult.Metrics.CompletionTokens
	chunkMetrics.TotalTokens += tocResult.Metrics.TotalTokens
	chunkMetrics.LLMCalls += tocResult.Metrics.LLMCalls
	chunkMetrics.PagesRead += tocResult.Metrics.PagesRead

	return sections, chunkMetrics, nil
}

func (s *Service) fail(ctx context.Context, docID string, err error) {
	if updateErr := s.docRepo.UpdateStatus(ctx, docID, domain.StatusError, err.Error()); updateErr != nil {
		s.logger.Error("failed to update error status", "doc_id", docID, "error", updateErr)
	}
}
