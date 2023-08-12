package tarantool

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/sku4/ad-notifier/model"
	clientModel "github.com/sku4/ad-parser/pkg/ad/model"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
)

func (ad *Ad) Take(ctx context.Context) (*clientModel.AdTnt, error) {
	call := tarantool.NewCallRequest("box.space.ad:take").Args([]interface{}{}).Context(ctx)
	resp, err := ad.conn.Do(call, pool.RW).Get()
	if err != nil {
		return nil, errors.Wrap(err, "take: call")
	}

	var adsTnt []*clientModel.AdTnt
	err = mapstructure.Decode(resp.Data, &adsTnt)
	if err != nil {
		return nil, errors.Wrap(err, "take: decode")
	}

	if len(adsTnt) > 0 {
		return adsTnt[0], nil
	}

	return nil, model.ErrQueueIsEmpty
}

func (ad *Ad) Bury(ctx context.Context, id uint64) error {
	_ = ctx

	call := tarantool.NewCallRequest("box.space.ad:bury").Args([]interface{}{id})
	_, err := ad.conn.Do(call, pool.RW).Get()
	if err != nil {
		return errors.Wrap(err, "bury: call")
	}

	return nil
}

func (ad *Ad) Ack(ctx context.Context, id uint64) error {
	_ = ctx

	call := tarantool.NewCallRequest("box.space.ad:ack").Args([]interface{}{id})
	_, err := ad.conn.Do(call, pool.RW).Get()
	if err != nil {
		return errors.Wrap(err, "ack: call")
	}

	return nil
}

func (ad *Ad) Release(ctx context.Context, id uint64) error {
	_ = ctx

	call := tarantool.NewCallRequest("box.space.ad:release").Args([]interface{}{id})
	_, err := ad.conn.Do(call, pool.RW).Get()
	if err != nil {
		return errors.Wrap(err, "release: call")
	}

	return nil
}
