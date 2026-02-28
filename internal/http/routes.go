package http

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(mux *http.ServeMux, h *Handler) {

	// ===== USER ROUTES =====

	// Exact match: /users
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			http.NotFound(w, r)
			return
		}
		h.users(w, r)
	})

	// Prefix match: /users/{id}
	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/" {
			http.NotFound(w, r)
			return
		}
		h.userByID(w, r)
	})

	// ===== HEALTH =====
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/ready", h.ready)

	// ===== METRICS =====
	mux.Handle("/metrics", promhttp.Handler())
}
