package rest

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/sku4/ad-notifier/configs"
)

func NewServer(ctx context.Context, handler http.Handler) *http.Server {
	cfg := configs.Get(ctx)
	return &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Rest.Port),
		Handler:        handler,
		MaxHeaderBytes: cfg.Rest.MaxHeaderBytes, // 1 MB
		ReadTimeout:    cfg.Rest.ReadTimeout,
		WriteTimeout:   cfg.Rest.WriteTimeout,
		BaseContext: func(n net.Listener) context.Context {
			return ctx
		},
	}
}
