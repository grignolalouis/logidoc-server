package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/logidoc/logidoc-server/internal/core/domain"
	"github.com/logidoc/logidoc-server/internal/core/port"
	"github.com/logidoc/logidoc-server/internal/core/service/indexer"
)

var _ port.DocumentService = (*DocumentService)(nil)

type DocumentService struct {
	docRepo   port.DocumentRepository
	indexRepo port.IndexRepository
	fileStore port.FileStore
	indexer   *indexer.Service
	logger    *slog.Logger
}

func NewDocumentService(
	docRepo port.DocumentRepository,
	indexRepo port.IndexRepository,
	fileStore port.FileStore,
	indexer *indexer.Service,
	logger *slog.Logger,
) *DocumentService {
	return &DocumentService{
		docRepo:   docRepo,
		indexRepo: indexRepo,
		fileStore: fileStore,
		indexer:   indexer,
		logger:    logger,
	}
}

// Upload stores the file and creates a document record. Does not trigger indexation.
func (s *DocumentService) Upload(ctx context.Context, filename string, file io.Reader) (*domain.Document, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	doc := &domain.Document{
		ID:        uuid.NewString(),
		Name:      filename,
		Status:    domain.StatusUploaded,
		CreatedAt: time.Now(),
	}

	if err := s.docRepo.Save(ctx, doc); err != nil {
		return nil, fmt.Errorf("save document: %w", err)
	}

	if err := s.fileStore.Save(ctx, doc.ID, data); err != nil {
		if delErr := s.docRepo.Delete(ctx, doc.ID); delErr != nil {
			s.logger.Error("failed to clean up document after file save error", "doc_id", doc.ID, "error", delErr)
		}
		return nil, fmt.Errorf("save file: %w", err)
	}

	s.logger.Info("document uploaded", "doc_id", doc.ID, "name", filename, "size", len(data))
	return doc, nil
}

// Index triggers async indexation for an uploaded document.
func (s *DocumentService) Index(ctx context.Context, id string) error {
	doc, err := s.docRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if doc.Status != domain.StatusUploaded && doc.Status != domain.StatusError {
		return fmt.Errorf("document %s is %s, cannot index", id, doc.Status)
	}

	data, err := s.fileStore.Load(ctx, id)
	if err != nil {
		return fmt.Errorf("load file: %w", err)
	}

	// Detached context — indexation runs independently of the HTTP request
	go func() {
		indexCtx := context.WithoutCancel(ctx)
		if err := s.indexer.Index(indexCtx, doc.ID, doc.Name, data); err != nil {
			s.logger.Error("indexation failed", "doc_id", doc.ID, "error", err)
		}
	}()

	return nil
}

func (s *DocumentService) Get(ctx context.Context, id string) (*domain.Document, error) {
	return s.docRepo.FindByID(ctx, id)
}

func (s *DocumentService) List(ctx context.Context) ([]domain.Document, error) {
	all, err := s.docRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	var docs []domain.Document
	for _, doc := range all {
		if doc.Status == domain.StatusReady {
			docs = append(docs, doc)
		}
	}
	return docs, nil
}

func (s *DocumentService) ListAll(ctx context.Context) ([]domain.Document, error) {
	return s.docRepo.FindAll(ctx)
}

func (s *DocumentService) Delete(ctx context.Context, id string) error {
	if _, err := s.docRepo.FindByID(ctx, id); err != nil {
		return err
	}
	if err := s.indexRepo.Delete(ctx, id); err != nil {
		s.logger.Warn("failed to delete index", "doc_id", id, "error", err)
	}
	if err := s.fileStore.Delete(ctx, id); err != nil {
		s.logger.Warn("failed to delete file", "doc_id", id, "error", err)
	}
	return s.docRepo.Delete(ctx, id)
}
