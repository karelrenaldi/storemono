package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// HealthCheck is a standard, simple health check
type HealthCheck struct{}

// AddRoutes adds the routers for this API to the provided router (or subrouter)
func (h *HealthCheck) AddRoutes(router *mux.Router) {
	router.HandleFunc("/health", h.handler).Methods("GET")
}

func (h *HealthCheck) handler(resp http.ResponseWriter, _ *http.Request) {
	_, _ = resp.Write([]byte(`OK`))
}
