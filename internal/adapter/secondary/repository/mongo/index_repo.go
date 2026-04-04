package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/logidoc/logidoc-server/internal/core/domain"
)

type indexModel struct {
	DocID   string      `bson:"_id"`
	Tree    []nodeModel `bson:"tree"`
	Version int         `bson:"version"`
}

type nodeModel struct {
	ID        string      `bson:"id"`
	Title     string      `bson:"title"`
	Summary   string      `bson:"summary"`
	Text      string      `bson:"text"`
	StartPage int         `bson:"start_page,omitempty"`
	EndPage   int         `bson:"end_page,omitempty"`
	Children  []nodeModel `bson:"children,omitempty"`
}

type IndexRepo struct {
	col *mongo.Collection
}

func NewIndexRepo(conn *Connection) *IndexRepo {
	return &IndexRepo{col: conn.Database().Collection("indexes")}
}

func (r *IndexRepo) Save(ctx context.Context, index *domain.Index) error {
	m := toIndexModel(index)
	_, err := r.col.ReplaceOne(ctx,
		bson.M{"_id": index.DocID},
		m,
		options.Replace().SetUpsert(true),
	)
	return err
}

func (r *IndexRepo) FindByDocID(ctx context.Context, docID string) (*domain.Index, error) {
	var m indexModel
	if err := r.col.FindOne(ctx, bson.M{"_id": docID}).Decode(&m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrIndexNotFound(docID)
		}
		return nil, err
	}
	return fromIndexModel(&m), nil
}

func (r *IndexRepo) Delete(ctx context.Context, docID string) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": docID})
	return err
}

func toIndexModel(idx *domain.Index) *indexModel {
	return &indexModel{
		DocID:   idx.DocID,
		Tree:    toNodeModels(idx.Tree),
		Version: idx.Version,
	}
}

func toNodeModels(nodes []domain.Node) []nodeModel {
	result := make([]nodeModel, len(nodes))
	for i, n := range nodes {
		result[i] = nodeModel{
			ID:        n.ID,
			Title:     n.Title,
			Summary:   n.Summary,
			Text:      n.Text,
			StartPage: n.StartPage,
			EndPage:   n.EndPage,
			Children:  toNodeModels(n.Children),
		}
	}
	return result
}

func fromIndexModel(m *indexModel) *domain.Index {
	return &domain.Index{
		DocID:   m.DocID,
		Tree:    fromNodeModels(m.Tree),
		Version: m.Version,
	}
}

func fromNodeModels(models []nodeModel) []domain.Node {
	result := make([]domain.Node, len(models))
	for i, m := range models {
		result[i] = domain.Node{
			ID:        m.ID,
			Title:     m.Title,
			Summary:   m.Summary,
			Text:      m.Text,
			StartPage: m.StartPage,
			EndPage:   m.EndPage,
			Children:  fromNodeModels(m.Children),
		}
	}
	return result
}
