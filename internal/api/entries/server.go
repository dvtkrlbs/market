package entries

import (
	"context"
	"errors"

	"github.com/dvtkrlbs/market/internal/api/common"
	"github.com/dvtkrlbs/market/internal/bookkeeping"
	"github.com/dvtkrlbs/market/internal/ent"
)

type Server struct {
	bookkeeping *bookkeeping.Service
}

var _ StrictServerInterface = (*Server)(nil)

func NewServer(service *bookkeeping.Service) *Server {
	return &Server{bookkeeping: service}
}

func (s *Server) ListEntries(ctx context.Context, _ ListEntriesRequestObject) (ListEntriesResponseObject, error) {
	items, err := s.bookkeeping.List(ctx)
	if err != nil {
		return nil, err
	}

	response := make(ListEntries200JSONResponse, 0, len(items))
	for _, item := range items {
		response = append(response, entryResponse(item))
	}
	return response, nil
}

func (s *Server) CreateEntry(ctx context.Context, request CreateEntryRequestObject) (CreateEntryResponseObject, error) {
	if request.Body == nil {
		return CreateEntry400JSONResponse(common.Error{Message: "request body is required"}), nil
	}

	created, err := s.bookkeeping.Create(ctx, bookkeeping.NewEntry{
		Description: request.Body.Description,
		AmountCents: request.Body.AmountCents,
		OccurredAt:  request.Body.OccurredAt,
	})
	if err != nil {
		if errors.Is(err, bookkeeping.ErrDescriptionRequired) {
			return CreateEntry400JSONResponse(common.Error{Message: err.Error()}), nil
		}
		return nil, err
	}
	return CreateEntry201JSONResponse(entryResponse(created)), nil
}

func entryResponse(item *ent.Entry) Entry {
	return Entry{
		ID:          item.ID,
		Description: item.Description,
		AmountCents: item.AmountCents,
		OccurredAt:  item.OccurredAt,
		CreatedAt:   item.CreatedAt,
	}
}
