package watcher

import (
	"github.com/pkg/errors"
	"github.com/sku4/ad-notifier/model"
	"github.com/sku4/ad-parser/pkg/logger"
	"github.com/tarantool/go-tarantool/v2"
)

func (w *Watcher) EventNewAd(event tarantool.WatchEvent) {
	_ = event
	log := logger.Get()

	err := w.services.Notifier.Run(w.ctx)
	if err != nil && !errors.Is(err, model.ErrNotifierIsRunning) {
		log.Errorf("error notifier run: %s", err)
	}
}
