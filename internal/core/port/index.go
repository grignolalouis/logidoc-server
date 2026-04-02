package port

import (
	"context"

	"github.com/logidoc/logidoc-server/internal/core/domain"
)

// IndexRepository handles persistence of document indexes.
type IndexRepository interface {
	Save(ctx context.Context, index *domain.Index) error
	FindByDocID(ctx context.Context, docID string) (*domain.Index, error)
	Delete(ctx context.Context, docID string) error
}
