package health

import "context"

type Server struct{}

var _ StrictServerInterface = (*Server)(nil)

func NewServer() *Server {
	return &Server{}
}

func (s *Server) GetHealth(context.Context, GetHealthRequestObject) (GetHealthResponseObject, error) {
	return GetHealth200JSONResponse{Status: "ok"}, nil
}
