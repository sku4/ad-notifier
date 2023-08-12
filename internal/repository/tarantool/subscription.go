package tarantool

import (
	"context"
)

func (ad *Ad) Filter(ctx context.Context, fields map[string]any) ([]int64, error) {
	return ad.client.SubscriptionFilter(ctx, fields)
}
