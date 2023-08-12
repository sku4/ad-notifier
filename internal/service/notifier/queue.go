package notifier

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"

	dec "github.com/shopspring/decimal"
	"github.com/sku4/ad-notifier/configs"
	"github.com/sku4/ad-notifier/internal/repository"
	"github.com/sku4/ad-notifier/internal/sender"
	"github.com/sku4/ad-notifier/model"
	clientModel "github.com/sku4/ad-parser/pkg/ad/model"
	"github.com/sku4/ad-parser/pkg/ad/street"
	"github.com/sku4/ad-parser/pkg/logger"
)

const (
	fieldsCount = 7
	powerOfTwo  = 2
	formatBase  = 2
)

type Queue struct {
	repos   *repository.Repository
	senders *sender.Sender
	rw      *sync.RWMutex
	cntTake int
	cntAck  int
	cntBury int
}

func NewQueue(repos *repository.Repository, senders *sender.Sender) *Queue {
	return &Queue{
		repos:   repos,
		senders: senders,
		rw:      &sync.RWMutex{},
	}
}

func (q *Queue) Run(ctx context.Context) {
	log := logger.Get()

	cntTake := 0
	cntAck := 0
	cntBury := 0

	defer func() {
		q.rw.Lock()
		q.cntTake += cntTake
		q.cntAck += cntAck
		q.cntBury += cntBury
		q.rw.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		adTnt, err := q.repos.Queue.Take(ctx)
		if err != nil {
			if !errors.Is(err, model.ErrQueueIsEmpty) {
				log.Errorf("error queue take: %s", err)
			}
			return
		}
		cntTake++

		err = q.process(ctx, adTnt)
		if err != nil {
			log.Errorf("error queue process id %d: %s", adTnt.ID, err)
			errBury := q.repos.Queue.Bury(ctx, adTnt.ID)
			if errBury != nil {
				log.Errorf("error queue bury id %d: %s", adTnt.ID, errBury)
				return
			}
			cntBury++
		} else {
			errAck := q.repos.Queue.Ack(ctx, adTnt.ID)
			if errAck != nil {
				log.Errorf("error queue ack id %d: %s", adTnt.ID, errAck)
				return
			}
			cntAck++
		}
	}
}

func (q *Queue) process(ctx context.Context, adTnt *clientModel.AdTnt) error {
	if adTnt == nil {
		return fmt.Errorf("process: ad is nil: %w", model.ErrNotFound)
	}

	wg := &sync.WaitGroup{}
	cfg := configs.Get(ctx)

	adTntFields := make(map[string]any, fieldsCount)
	sortFields := make([]string, 0, fieldsCount)

	features := cfg.Features
	featuresSet := make(map[string]struct{})
	for _, f := range features {
		featuresSet[f] = struct{}{}
	}

	if hasFeature("street_id", featuresSet) && adTnt.StreetID != nil && *adTnt.StreetID > 0 {
		adTntFields["street_id"] = *adTnt.StreetID
		sortFields = append(sortFields, "street_id")
	}

	if hasFeature("house", featuresSet) && adTnt.House != nil && *adTnt.House != "" {
		adTntFields["house"] = *adTnt.House
		sortFields = append(sortFields, "house")
	}

	decZero := dec.New(0, 0)
	if hasFeature("price", featuresSet) {
		if adTnt.Price != nil && adTnt.Price.GreaterThan(decZero) {
			adTntFields["price"] = adTnt.Price
		} else {
			adTntFields["price"] = -1
		}
		sortFields = append(sortFields, "price")
	}

	if hasFeature("price_m2", featuresSet) {
		if adTnt.PriceM2 != nil && adTnt.PriceM2.GreaterThan(decZero) {
			adTntFields["price_m2"] = adTnt.PriceM2
		} else {
			adTntFields["price_m2"] = -1
		}
		sortFields = append(sortFields, "price_m2")
	}

	if hasFeature("rooms", featuresSet) {
		if adTnt.Rooms != nil && *adTnt.Rooms > 0 {
			adTntFields["rooms"] = *adTnt.Rooms
		} else {
			adTntFields["rooms"] = -1
		}
		sortFields = append(sortFields, "rooms")
	}

	if hasFeature("floor", featuresSet) {
		if adTnt.Floor != nil && *adTnt.Floor > 0 {
			adTntFields["floor"] = *adTnt.Floor
		} else {
			adTntFields["floor"] = -1
		}
		sortFields = append(sortFields, "floor")
	}

	if hasFeature("year", featuresSet) {
		if adTnt.Year != nil && *adTnt.Year > 0 {
			adTntFields["year"] = *adTnt.Year
		} else {
			adTntFields["year"] = -1
		}
		sortFields = append(sortFields, "year")
	}

	if hasFeature("m2_main", featuresSet) {
		if adTnt.M2Main != nil && *adTnt.M2Main > 0 {
			adTntFields["m2_main"] = *adTnt.M2Main
		} else {
			adTntFields["m2_main"] = -1
		}
		sortFields = append(sortFields, "m2_main")
	}

	cnt := int64(math.Pow(powerOfTwo, float64(len(sortFields))))
	tgIdsSet := make(map[int64]struct{})
	errs := make([]error, 0)

	wg.Add(cfg.Notifier.WorkerSubscriptionCount)
	for w := 0; w < cfg.Notifier.WorkerSubscriptionCount; w++ {
		go func(ctx context.Context, wg *sync.WaitGroup, adTntFields map[string]any, sortFields []string) {
			defer wg.Done()

			for {
				q.rw.Lock()
				if cnt == 1 {
					q.rw.Unlock()
					return
				}
				cnt--
				c := cnt
				q.rw.Unlock()

				fields := make(map[string]any)
				bb := []byte(strconv.FormatInt(c, formatBase))
				shift := len(sortFields) - len(bb)
				for i := len(bb) - 1; i >= 0; i-- {
					if bb[i] == '1' {
						fieldName := sortFields[i+shift]
						fields[fieldName] = adTntFields[fieldName]
					}
				}

				filterIds, err := q.repos.Subscription.Filter(ctx, fields)
				if err != nil {
					q.rw.Lock()
					errs = append(errs, fmt.Errorf("subscription filter: %w fields %v", err, fields))
					q.rw.Unlock()
					continue
				}

				q.rw.Lock()
				for _, id := range filterIds {
					tgIdsSet[id] = struct{}{}
				}
				q.rw.Unlock()
			}
		}(ctx, wg, adTntFields, sortFields)
	}

	wg.Wait()

	tgIds := make([]int64, 0, len(tgIdsSet))
	for id := range tgIdsSet {
		tgIds = append(tgIds, id)
	}

	if len(tgIds) > 0 {
		// get street name
		var streetExt *street.Ext
		var errStreet error
		if adTnt.StreetID != nil && *adTnt.StreetID > 0 {
			streetExt, errStreet = q.repos.Street.GetStreet(ctx, *adTnt.StreetID)
			if errStreet != nil {
				errs = append(errs, errStreet)
				return errors.Join(errs...)
			}
		}

		// send notifications about new ad
		tgErrs := q.senders.Telegram.NotifyNewAd(ctx, adTnt, streetExt, tgIds)
		errs = append(errs, tgErrs...)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func hasFeature(feature string, features map[string]struct{}) bool {
	_, ok := features[feature]
	return ok
}
