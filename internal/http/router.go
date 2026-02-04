package http

import (
	"github.com/go-chi/chi/v5"
)

func Router(h *Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Post("/transactions", h.CreateTransaction)
	r.Get("/transactions", h.ListTransactions)
	r.Get("/transactions/{hash}", h.GetTransaction)
	r.Get("/stats", h.GetStats)

	return r
}
