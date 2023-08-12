package watcher

import (
	"context"

	"github.com/sku4/ad-notifier/internal/service"
	"github.com/sku4/ad-parser/pkg/ad/model"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
)

type Watcher struct {
	ctx      context.Context
	services *service.Service
	watchers []tarantool.Watcher
}

func NewWatcher(ctx context.Context, services *service.Service) *Watcher {
	return &Watcher{
		ctx:      ctx,
		services: services,
		watchers: make([]tarantool.Watcher, 0),
	}
}

func (w *Watcher) Register() error {
	watchersMatch := map[string]tarantool.WatchCallback{
		model.EventNewAd: w.EventNewAd,
	}
	for key, callback := range watchersMatch {
		watcher, err := w.services.Watcher.Register(key, callback, pool.RO)
		if err != nil {
			return err
		}
		w.watchers = append(w.watchers, watcher)
	}

	return nil
}

func (w *Watcher) Unregister() {
	for _, watcher := range w.watchers {
		watcher.Unregister()
	}
}
