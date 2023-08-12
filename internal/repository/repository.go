package repository

import (
	"context"

	tnt "github.com/sku4/ad-notifier/internal/repository/tarantool"
	clientModel "github.com/sku4/ad-parser/pkg/ad/model"
	"github.com/sku4/ad-parser/pkg/ad/street"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
)

//go:generate mockgen -source=repository.go -destination=mocks/repository.go

type Queue interface {
	Take(ctx context.Context) (*clientModel.AdTnt, error)
	Ack(ctx context.Context, id uint64) error
	Release(ctx context.Context, id uint64) error
	Bury(ctx context.Context, id uint64) error
}

type Subscription interface {
	Filter(ctx context.Context, fields map[string]any) ([]int64, error)
}

type Watcher interface {
	Register(string, tarantool.WatchCallback, pool.Mode) (tarantool.Watcher, error)
}

type Street interface {
	GetStreet(ctx context.Context, id uint64) (*street.Ext, error)
}

type Repository struct {
	Queue
	Subscription
	Watcher
	Street
}

func NewRepository(conn pool.Pooler) *Repository {
	ad := tnt.NewAd(conn)
	return &Repository{
		Queue:        ad,
		Subscription: ad,
		Watcher:      ad,
		Street:       ad,
	}
}
