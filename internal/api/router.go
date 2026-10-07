package api

import (
	"net/http"

	"github.com/dvtkrlbs/market/internal/api/entries"
	"github.com/dvtkrlbs/market/internal/api/health"
	"github.com/dvtkrlbs/market/internal/api/orders"
	"github.com/dvtkrlbs/market/internal/bookkeeping"
)

func NewRouter(service *bookkeeping.Service) http.Handler {
	mux := http.NewServeMux()

	health.HandlerFromMux(
		health.NewStrictHandler(health.NewServer(), nil),
		mux,
	)
	entries.HandlerFromMux(
		entries.NewStrictHandler(entries.NewServer(service), nil),
		mux,
	)
	orders.HandlerFromMux(
		orders.NewStrictHandler(orders.NewServer(), nil),
		mux,
	)

	return mux
}
