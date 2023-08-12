package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) InitRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)

	r.Handle("/metrics", promhttp.Handler())

	return r
}
