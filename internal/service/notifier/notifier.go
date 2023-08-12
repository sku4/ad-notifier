package notifier

import (
	"context"
	"sync"

	"github.com/sku4/ad-notifier/configs"
	"github.com/sku4/ad-notifier/internal/repository"
	"github.com/sku4/ad-notifier/internal/sender"
	"github.com/sku4/ad-notifier/model"
	"github.com/sku4/ad-parser/pkg/logger"
)

//go:generate mockgen -source=notifier.go -destination=mocks/notifier.go

type Service struct {
	repos   *repository.Repository
	senders *sender.Sender
	wg      *sync.WaitGroup
	rw      *sync.RWMutex
	running bool
}

func NewService(repos *repository.Repository, senders *sender.Sender) *Service {
	return &Service{
		repos:   repos,
		senders: senders,
		wg:      &sync.WaitGroup{},
		rw:      &sync.RWMutex{},
	}
}

func (s *Service) Run(ctx context.Context) error {
	s.rw.RLock()
	if s.running {
		s.rw.RUnlock()
		return model.ErrNotifierIsRunning
	}
	s.rw.RUnlock()

	s.rw.Lock()
	s.running = true
	s.rw.Unlock()

	log := logger.Get()
	cfg := configs.Get(ctx)

	s.wg.Add(cfg.Notifier.WorkerQueueCount)
	queue := NewQueue(s.repos, s.senders)
	for w := 0; w < cfg.Notifier.WorkerQueueCount; w++ {
		go func(ctx context.Context, wg *sync.WaitGroup, queue *Queue) {
			defer wg.Done()
			queue.Run(ctx)
		}(ctx, s.wg, queue)
	}

	log.Info("Notifier worker running")
	s.wg.Wait()
	s.rw.Lock()
	s.running = false
	s.rw.Unlock()
	log.Infof("Queue stat: take %d, ack %d, bury %d", queue.cntTake, queue.cntAck, queue.cntBury)
	log.Info("Notifier worker finished")

	return nil
}

func (s *Service) Shutdown() error {
	s.wg.Wait()

	return nil
}
