package stores

import (
	"context"
	"errors"

	"github.com/dvtkrlbs/market/internal/api/common"
	requesthelper "github.com/dvtkrlbs/market/internal/api/request"
	"github.com/dvtkrlbs/market/internal/ent"
	authservice "github.com/dvtkrlbs/market/internal/services/auth"
	storeservice "github.com/dvtkrlbs/market/internal/services/stores"
)

type service interface {
	List(context.Context, int, string) ([]*ent.Store, error)
	Create(context.Context, int, string) (*ent.Store, error)
	Get(context.Context, int, int) (*ent.Store, error)
	Update(context.Context, int, int, storeservice.Update) (*ent.Store, error)
	Delete(context.Context, int, int) error
	Approve(context.Context, int) (*ent.Store, error)
}

type Server struct {
	stores service
	admins requesthelper.AdminChecker
}

var _ StrictServerInterface = (*Server)(nil)

func NewServer(stores service, admins requesthelper.AdminChecker) *Server {
	return &Server{stores: stores, admins: admins}
}

func (s *Server) ApproveStore(ctx context.Context, input ApproveStoreRequestObject) (ApproveStoreResponseObject, error) {
	if err := requesthelper.RequireAdmin(ctx, s.admins); err != nil {
		switch {
		case errors.Is(err, requesthelper.ErrUnauthenticated):
			return ApproveStore401JSONResponse{StoreUnauthorizedJSONResponse{Message: err.Error()}}, nil
		case errors.Is(err, requesthelper.ErrAdminRequired):
			return ApproveStore403JSONResponse{StoreForbiddenJSONResponse{Message: err.Error()}}, nil
		default:
			return nil, err
		}
	}
	item, err := s.stores.Approve(ctx, input.StoreId)
	if errors.Is(err, storeservice.ErrNotFound) {
		return ApproveStore404JSONResponse{StoreNotFoundJSONResponse{Message: err.Error()}}, nil
	}
	if err != nil {
		return nil, err
	}
	return ApproveStore200JSONResponse(storeResponse(item)), nil
}

func authorize(ctx context.Context, userID int) int {
	principal, ok := authservice.PrincipalFromContext(ctx)
	if !ok || principal.User == nil {
		return 401
	}
	if principal.User.ID != userID {
		return 403
	}
	return 0
}

func storeResponse(item *ent.Store) common.Store {
	return common.Store{ID: item.ID, Name: item.Name, Approved: item.Approved}
}

func (s *Server) ListStores(ctx context.Context, request ListStoresRequestObject) (ListStoresResponseObject, error) {
	switch authorize(ctx, request.UserId) {
	case 401:
		return ListStores401JSONResponse{StoreUnauthorizedJSONResponse{Message: "authentication required"}}, nil
	case 403:
		return ListStores403JSONResponse{StoreForbiddenJSONResponse{Message: "user mismatch"}}, nil
	}
	mode := "shop"
	if request.Params.Context != nil {
		mode = string(*request.Params.Context)
	}
	items, err := s.stores.List(ctx, request.UserId, mode)
	if err != nil {
		return nil, err
	}
	response := make(ListStores200JSONResponse, 0, len(items))
	for _, item := range items {
		response = append(response, storeResponse(item))
	}
	return response, nil
}

func (s *Server) CreateStore(ctx context.Context, request CreateStoreRequestObject) (CreateStoreResponseObject, error) {
	switch authorize(ctx, request.UserId) {
	case 401:
		return CreateStore401JSONResponse{StoreUnauthorizedJSONResponse{Message: "authentication required"}}, nil
	case 403:
		return CreateStore403JSONResponse{StoreForbiddenJSONResponse{Message: "user mismatch"}}, nil
	}
	if request.Body == nil {
		return CreateStore400JSONResponse{StoreBadRequestJSONResponse{Message: "request body is required"}}, nil
	}
	item, err := s.stores.Create(ctx, request.UserId, request.Body.Name)
	if errors.Is(err, storeservice.ErrNameRequired) {
		return CreateStore400JSONResponse{StoreBadRequestJSONResponse{Message: err.Error()}}, nil
	}
	if err != nil {
		return nil, err
	}
	return CreateStore201JSONResponse(storeResponse(item)), nil
}

func (s *Server) GetStore(ctx context.Context, request GetStoreRequestObject) (GetStoreResponseObject, error) {
	switch authorize(ctx, request.UserId) {
	case 401:
		return GetStore401JSONResponse{StoreUnauthorizedJSONResponse{Message: "authentication required"}}, nil
	case 403:
		return GetStore403JSONResponse{StoreForbiddenJSONResponse{Message: "user mismatch"}}, nil
	}
	item, err := s.stores.Get(ctx, request.UserId, request.StoreId)
	if errors.Is(err, storeservice.ErrNotFound) {
		return GetStore404JSONResponse{StoreNotFoundJSONResponse{Message: err.Error()}}, nil
	}
	if err != nil {
		return nil, err
	}
	return GetStore200JSONResponse(storeResponse(item)), nil
}

func (s *Server) UpdateStore(ctx context.Context, request UpdateStoreRequestObject) (UpdateStoreResponseObject, error) {
	switch authorize(ctx, request.UserId) {
	case 401:
		return UpdateStore401JSONResponse{StoreUnauthorizedJSONResponse{Message: "authentication required"}}, nil
	case 403:
		return UpdateStore403JSONResponse{StoreForbiddenJSONResponse{Message: "user mismatch"}}, nil
	}
	if request.Body == nil {
		return UpdateStore400JSONResponse{StoreBadRequestJSONResponse{Message: "request body is required"}}, nil
	}
	if request.Body.Approved != nil {
		if err := requesthelper.RequireAdmin(ctx, s.admins); err != nil {
			if errors.Is(err, requesthelper.ErrAdminRequired) {
				return UpdateStore403JSONResponse{StoreForbiddenJSONResponse{Message: err.Error()}}, nil
			}
			return nil, err
		}
	}
	item, err := s.stores.Update(ctx, request.UserId, request.StoreId, storeservice.Update{Name: request.Body.Name, Approved: request.Body.Approved})
	switch {
	case errors.Is(err, storeservice.ErrNameRequired), errors.Is(err, storeservice.ErrEmptyUpdate):
		return UpdateStore400JSONResponse{StoreBadRequestJSONResponse{Message: err.Error()}}, nil
	case errors.Is(err, storeservice.ErrForbidden):
		return UpdateStore403JSONResponse{StoreForbiddenJSONResponse{Message: err.Error()}}, nil
	case errors.Is(err, storeservice.ErrNotFound):
		return UpdateStore404JSONResponse{StoreNotFoundJSONResponse{Message: err.Error()}}, nil
	case err != nil:
		return nil, err
	}
	return UpdateStore200JSONResponse(storeResponse(item)), nil
}

func (s *Server) DeleteStore(ctx context.Context, request DeleteStoreRequestObject) (DeleteStoreResponseObject, error) {
	switch authorize(ctx, request.UserId) {
	case 401:
		return DeleteStore401JSONResponse{StoreUnauthorizedJSONResponse{Message: "authentication required"}}, nil
	case 403:
		return DeleteStore403JSONResponse{StoreForbiddenJSONResponse{Message: "user mismatch"}}, nil
	}
	err := s.stores.Delete(ctx, request.UserId, request.StoreId)
	switch {
	case errors.Is(err, storeservice.ErrForbidden):
		return DeleteStore403JSONResponse{StoreForbiddenJSONResponse{Message: err.Error()}}, nil
	case errors.Is(err, storeservice.ErrNotFound):
		return DeleteStore404JSONResponse{StoreNotFoundJSONResponse{Message: err.Error()}}, nil
	case err != nil:
		return nil, err
	}
	return DeleteStore204Response{}, nil
}
