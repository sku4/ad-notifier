package tarantool

import (
	"context"

	"github.com/sku4/ad-parser/pkg/ad/street"
)

func (ad *Ad) GetStreet(ctx context.Context, id uint64) (*street.Ext, error) {
	return ad.client.StreetGet(ctx, id)
}
