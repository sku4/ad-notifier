package tarantool

import (
	client "github.com/sku4/ad-parser/pkg/ad"
	"github.com/tarantool/go-tarantool/v2/pool"
)

type Ad struct {
	conn   pool.Pooler
	client *client.Client
}

func NewAd(conn pool.Pooler) *Ad {
	clientTnt := client.NewClient(conn)
	return &Ad{
		conn:   conn,
		client: clientTnt,
	}
}
