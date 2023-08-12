package watcher

import (
	"github.com/sku4/ad-notifier/internal/repository"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
)

type Watcher struct {
	repos *repository.Repository
}

func NewWatcher(repos *repository.Repository) *Watcher {
	return &Watcher{
		repos: repos,
	}
}

func (w Watcher) Register(key string, callback tarantool.WatchCallback, mode pool.Mode) (tarantool.Watcher, error) {
	return w.repos.Watcher.Register(key, callback, mode)
}
