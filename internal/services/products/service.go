package products

import (
	"context"

	"github.com/dvtkrlbs/market/internal/ent"
	"github.com/dvtkrlbs/market/internal/ent/store"
)

type Service struct{ client *ent.Client }

func New(client *ent.Client) *Service { return &Service{client: client} }

func (s *Service) List(ctx context.Context) ([]*ent.Product, error) {
	query := s.client.Product.Query().Order(ent.Asc(store.FieldID))

	return query.All(ctx)
}
