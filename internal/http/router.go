package http

import (
	"net/http"
	"github.com/go-chi/chi/v5"
)

func Router(h *Handlers) *chi.Mux {
	r := chi.NewRouter()

	// CORS middleware - MUST come first!
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow requests from any origin (for development)
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "3600")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Rate limiting middleware
	r.Use(RateLimitMiddleware)

	// Routes
	r.Post("/transactions", h.CreateTransaction)
	r.Get("/transactions", h.ListTransactions)
	r.Get("/transactions/{hash}", h.GetTransaction)
	r.Get("/stats", h.GetStats)

	return r
}
