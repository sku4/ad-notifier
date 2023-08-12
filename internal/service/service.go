package service

import (
	"context"

	"github.com/sku4/ad-notifier/internal/repository"
	"github.com/sku4/ad-notifier/internal/sender"
	"github.com/sku4/ad-notifier/internal/service/notifier"
	"github.com/sku4/ad-notifier/internal/service/watcher"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
)

//go:generate mockgen -source=service.go -destination=mocks/service.go

type Runner interface {
	Run(context.Context) error
	Shutdown() error
}

type Watcher interface {
	Register(string, tarantool.WatchCallback, pool.Mode) (tarantool.Watcher, error)
}

type Service struct {
	Notifier Runner
	Watcher
}

func NewService(repos *repository.Repository, senders *sender.Sender) *Service {
	return &Service{
		Notifier: notifier.NewService(repos, senders),
		Watcher:  watcher.NewWatcher(repos),
	}
}
