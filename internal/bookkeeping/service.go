package bookkeeping

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dvtkrlbs/market/internal/ent"
	"github.com/dvtkrlbs/market/internal/ent/entry"
)

var ErrDescriptionRequired = errors.New("description is required")

type NewEntry struct {
	Description string
	AmountCents int64
	OccurredAt  time.Time
}

type Service struct {
	client *ent.Client
}

func New(client *ent.Client) *Service {
	return &Service{client: client}
}

func (s *Service) Create(ctx context.Context, input NewEntry) (*ent.Entry, error) {
	description := strings.TrimSpace(input.Description)
	if description == "" {
		return nil, ErrDescriptionRequired
	}
	created, err := s.client.Entry.Create().
		SetDescription(description).
		SetAmountCents(input.AmountCents).
		SetOccurredAt(input.OccurredAt).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create entry: %w", err)
	}
	return created, nil
}

func (s *Service) List(ctx context.Context) ([]*ent.Entry, error) {
	entries, err := s.client.Entry.Query().
		Order(ent.Desc(entry.FieldOccurredAt), ent.Desc(entry.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	return entries, nil
}
