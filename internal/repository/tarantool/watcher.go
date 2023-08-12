package tarantool

import (
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
)

func (ad *Ad) Register(key string, callback tarantool.WatchCallback, mode pool.Mode) (tarantool.Watcher, error) {
	return ad.conn.NewWatcher(key, callback, mode)
}
