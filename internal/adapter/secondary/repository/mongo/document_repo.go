package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/logidoc/logidoc-server/internal/core/domain"
)

type documentModel struct {
	ID          string  `bson:"_id"`
	Name        string  `bson:"name"`
	Description string  `bson:"description,omitempty"`
	Status      string  `bson:"status"`
	PageCount   int     `bson:"page_count"`
	NodeCount   int     `bson:"node_count"`
	Error       string  `bson:"error,omitempty"`
	CreatedAt   int64   `bson:"created_at"`
	IndexedAt   *int64  `bson:"indexed_at,omitempty"`
}

// DocumentRepo implements port.DocumentRepository with MongoDB.
type DocumentRepo struct {
	col *mongo.Collection
}

// NewDocumentRepo creates a new DocumentRepo.
func NewDocumentRepo(conn *Connection) *DocumentRepo {
	return &DocumentRepo{col: conn.Database().Collection("documents")}
}

func (r *DocumentRepo) Save(ctx context.Context, doc *domain.Document) error {
	_, err := r.col.InsertOne(ctx, toDocModel(doc))
	return err
}

func (r *DocumentRepo) FindByID(ctx context.Context, id string) (*domain.Document, error) {
	var m documentModel
	if err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrDocumentNotFound
		}
		return nil, err
	}
	return fromDocModel(&m), nil
}

func (r *DocumentRepo) FindAll(ctx context.Context) ([]domain.Document, error) {
	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var models []documentModel
	if err := cursor.All(ctx, &models); err != nil {
		return nil, err
	}

	docs := make([]domain.Document, len(models))
	for i, m := range models {
		docs[i] = *fromDocModel(&m)
	}
	return docs, nil
}

func (r *DocumentRepo) UpdateStatus(ctx context.Context, id string, status domain.DocumentStatus, errMsg string) error {
	update := bson.M{"$set": bson.M{"status": string(status), "error": errMsg}}
	if status == domain.StatusReady {
		now := time.Now().UnixMilli()
		update["$set"].(bson.M)["indexed_at"] = now
	}
	_, err := r.col.UpdateByID(ctx, id, update)
	return err
}

func (r *DocumentRepo) Delete(ctx context.Context, id string) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func toDocModel(doc *domain.Document) *documentModel {
	m := &documentModel{
		ID:          doc.ID,
		Name:        doc.Name,
		Description: doc.Description,
		Status:      string(doc.Status),
		PageCount:   doc.PageCount,
		NodeCount:   doc.NodeCount,
		Error:       doc.Error,
		CreatedAt:   doc.CreatedAt.UnixMilli(),
	}
	if doc.IndexedAt != nil {
		ts := doc.IndexedAt.UnixMilli()
		m.IndexedAt = &ts
	}
	return m
}

func fromDocModel(m *documentModel) *domain.Document {
	doc := &domain.Document{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Status:      domain.DocumentStatus(m.Status),
		PageCount:   m.PageCount,
		NodeCount:   m.NodeCount,
		Error:       m.Error,
		CreatedAt:   time.UnixMilli(m.CreatedAt),
	}
	if m.IndexedAt != nil {
		t := time.UnixMilli(*m.IndexedAt)
		doc.IndexedAt = &t
	}
	return doc
}
