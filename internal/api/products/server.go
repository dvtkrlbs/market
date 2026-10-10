package products

import (
	"context"

	"github.com/dvtkrlbs/market/internal/api/common"
	"github.com/dvtkrlbs/market/internal/ent"
	productservice "github.com/dvtkrlbs/market/internal/services/products"
)

type Server struct {
	products *productservice.Service
}

var _ StrictServerInterface = (*Server)(nil)

func NewServer(service *productservice.Service) *Server {
	return &Server{products: service}
}

func (s *Server) ListProducts(ctx context.Context, request ListProductsRequestObject) (ListProductsResponseObject, error) {
	all, err := s.products.List(ctx)
	if err != nil {
		return nil, err
	}

	response := make(ListProducts200JSONResponse, 0, len(all))
	for _, item := range all {
		response = append(response, productResponse(item, string(*request.Params.Context)))
	}
	return response, nil
}

func productResponse(product *ent.Product, scope string) common.Product {
	base := common.Product{
		ID:   product.ID,
		Name: product.Name,
	}
	if scope != "shop" {
		base.Barcode = &product.Barcode
		base.CreatedAt = &product.CreatedAt
		base.DeletedAt = product.DeletedAt
	}

	return base
}
