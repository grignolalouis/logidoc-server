package index

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"trpc.group/trpc-go/trpc-agent-go/model"

	"os"

	"github.com/logidoc/logidoc-server/internal/core/domain"
	"github.com/logidoc/logidoc-server/internal/core/port"
)

type Config struct {
	MaxPagesPerNode        int
	EnableTableExtraction  bool
	EnableImageDescription bool
}

type Service struct {
	llm     model.Model
	vision  model.Model // nil = use llm for vision tasks
	docRepo port.DocumentRepository
	idxRepo port.IndexRepository
	cfg     Config
	logger  *slog.Logger
}

// NewService creates a new index service. If vision is nil, llm is used for vision tasks.
func NewService(llm model.Model, vision model.Model, docRepo port.DocumentRepository, idxRepo port.IndexRepository, cfg Config, logger *slog.Logger) *Service {
	if cfg.MaxPagesPerNode <= 0 {
		cfg.MaxPagesPerNode = 20
	}
	return &Service{llm: llm, vision: vision, docRepo: docRepo, idxRepo: idxRepo, cfg: cfg, logger: logger}
}

func (s *Service) visionModel() model.Model {
	if s.vision != nil {
		return s.vision
	}
	return s.llm
}

func (s *Service) Index(ctx context.Context, docID string, filename string, data []byte) error {
	start := time.Now()
	s.logger.Info("starting indexation", "doc_id", docID, "filename", filename, "size", len(data))

	if err := s.docRepo.UpdateStatus(ctx, docID, domain.StatusIndexing, ""); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	pages, err := ParseFile(filename, data)
	if err != nil {
		s.fail(ctx, docID, err)
		return fmt.Errorf("parse file: %w", err)
	}
	if pages.PDFPath != "" {
		defer os.Remove(pages.PDFPath)
	}
	s.logger.Info("parsed document", "doc_id", docID, "pages", pages.Count())

	// TODO: OCR for empty/scanned pages (VLM or tesseract)

	var metrics Metrics

	sections, structMetrics, err := s.detectStructure(ctx, pages)
	if err != nil {
		s.fail(ctx, docID, err)
		return err
	}
	metrics.merge(structMetrics)

	if len(sections) > 0 {
		before := sections[0].StartPage
		sections = CalibratePages(sections, pages)
		sections = VerifyCalibration(sections, pages, s.logger)
		if sections[0].StartPage != before {
			s.logger.Info("page offset detected", "doc_id", docID, "offset", sections[0].StartPage-before)
		}
	}

	tree := BuildTree(sections, pages.Count())
	tree = SubdivideLargeNodes(ctx, tree, pages, s.llm, s.logger, s.cfg.MaxPagesPerNode)

	if s.cfg.EnableTableExtraction && pages.PDFPath != "" {
		enrichTablesVLM(ctx, s.visionModel(), pages, &metrics, s.logger)
	}

	if s.cfg.EnableImageDescription && pages.PDFPath != "" {
		if err := enrichWithImages(ctx, s.visionModel(), pages, &metrics, s.logger); err != nil {
			s.logger.Warn("image enrichment failed", "doc_id", docID, "error", err)
		}
	}

	nodes := FillTextEnriched(tree, pages)

	if err := s.idxRepo.Save(ctx, &domain.Index{DocID: docID, Tree: nodes, Version: 1}); err != nil {
		s.fail(ctx, docID, err)
		return fmt.Errorf("save index: %w", err)
	}
	if err := s.docRepo.UpdateStatus(ctx, docID, domain.StatusReady, ""); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	metrics.Duration = time.Since(start)
	metrics.PagesTotal = pages.Count()
	metrics.SectionsFound = CountNodes(nodes)

	s.logger.Info("indexation complete",
		"doc_id", docID,
		"duration", metrics.Duration.Round(time.Millisecond),
		"pages", metrics.PagesTotal,
		"sections", metrics.SectionsFound,
		"pages_read", metrics.PagesRead,
		"agent_calls", metrics.AgentCalls,
		"agent_tokens", metrics.AgentTotalTokens,
		"vision_calls", metrics.VisionCalls,
		"vision_tokens", metrics.VisionTotalTokens,
		"total_tokens", metrics.TotalTokens(),
	)
	return nil
}

func (s *Service) detectStructure(ctx context.Context, pages *Pages) ([]FlatSection, *Metrics, error) {
	tocResult, err := DetectTOC(ctx, s.llm, pages)
	if err != nil {
		return nil, nil, err
	}
	if tocResult.Found {
		s.logger.Info("TOC detected", "sections", len(tocResult.Sections))
		return tocResult.Sections, &tocResult.Metrics, nil
	}

	s.logger.Info("no TOC detected, switching to sequential chunking")
	sections, chunkMetrics, err := ProcessNoTOC(ctx, s.llm, pages, s.logger)
	if err != nil {
		return nil, nil, err
	}

	chunkMetrics.merge(&tocResult.Metrics)
	return sections, chunkMetrics, nil
}

func (s *Service) fail(ctx context.Context, docID string, err error) {
	if updateErr := s.docRepo.UpdateStatus(ctx, docID, domain.StatusError, err.Error()); updateErr != nil {
		s.logger.Error("failed to update error status", "doc_id", docID, "error", updateErr)
	}
}
