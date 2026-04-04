package port

import (
	"context"
	"io"

	"github.com/logidoc/logidoc-server/internal/core/domain"
)

type DocumentService interface {
	Upload(ctx context.Context, filename string, file io.Reader) (*domain.Document, error)
	Index(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*domain.Document, error)
	List(ctx context.Context) ([]domain.Document, error)    // ready documents only
	ListAll(ctx context.Context) ([]domain.Document, error) // all documents
	Delete(ctx context.Context, id string) error
}

type DocumentRepository interface {
	Save(ctx context.Context, doc *domain.Document) error
	FindByID(ctx context.Context, id string) (*domain.Document, error)
	FindAll(ctx context.Context) ([]domain.Document, error)
	UpdateStatus(ctx context.Context, id string, status domain.DocumentStatus, errMsg string) error
	Delete(ctx context.Context, id string) error
}

type FileRepository interface {
	Save(ctx context.Context, id string, data []byte) error
	Load(ctx context.Context, id string) ([]byte, error)
	Delete(ctx context.Context, id string) error
}
