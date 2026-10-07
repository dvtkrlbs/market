package orders

import "context"

type Server struct{}

var _ StrictServerInterface = (*Server)(nil)

func NewServer() *Server {
	return &Server{}
}

func (s *Server) GetOrdersPlaceholder(context.Context, GetOrdersPlaceholderRequestObject) (GetOrdersPlaceholderResponseObject, error) {
	return GetOrdersPlaceholder200JSONResponse{Status: "placeholder"}, nil
}
