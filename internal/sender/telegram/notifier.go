package telegram

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sku4/ad-notifier/model"
	clientModel "github.com/sku4/ad-parser/pkg/ad/model"
	"github.com/sku4/ad-parser/pkg/ad/street"
	"github.com/sku4/ad-parser/pkg/logger"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	tmplTgNewAd = "tg_new_ad"
	secondCap   = 3
	m2Cap       = 3
	floorCap    = 2
	roundNumber = 100
	retryCount  = 2
)

var (
	sendAdTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "notifier",
		Subsystem: "tg_bot",
		Name:      "send_ad_total",
	}, []string{"status"})
)

func (t Telegram) NotifyNewAd(ctx context.Context, adTnt *clientModel.AdTnt,
	streetExt *street.Ext, tgIDs []int64) []error {
	log := logger.Get()
	errs := make([]error, 0)

	if adTnt == nil {
		errs = append(errs, fmt.Errorf("notify new ad: ad is nil: %w", model.ErrNotFound))
		return errs
	}

	addressParts := make([]string, 0)
	if streetExt != nil {
		addressParts = append(addressParts, streetExt.Street.Name)
		streetType := ""
		if streetExt.Type.Short != "" {
			streetType = streetExt.Type.Short + "."
		}
		if streetType != "" {
			addressParts = append(addressParts, streetType)
		}
	}
	if adTnt.House != nil && *adTnt.House != "" {
		addressParts = append(addressParts, *adTnt.House)
	}
	address := ""
	if len(addressParts) > 0 {
		address = fmt.Sprintf("*%s*", strings.Join(addressParts, " "))
	}
	if adTnt.Year != nil && *adTnt.Year > 0 {
		if address != "" {
			address += fmt.Sprintf("  (%d г.п.)", *adTnt.Year)
		} else {
			address = fmt.Sprintf("%d г.п.", *adTnt.Year)
		}
	}

	price := ""
	priceFloat := float64(0)
	if adTnt.Price != nil {
		priceFloat, _ = adTnt.Price.Float64()
	}
	priceM2Float := float64(0)
	if adTnt.PriceM2 != nil {
		priceM2Float, _ = adTnt.PriceM2.Float64()
	}
	if priceFloat > 0 {
		printFr := message.NewPrinter(language.French)
		price = printFr.Sprintf("*%.0f $*", priceFloat)

		if priceM2Float > 0 {
			price = fmt.Sprintf("%s   —   %s / м²", price, printFr.Sprintf("%.0f $", priceM2Float))
		}
	}

	second := make([]string, 0, secondCap)
	if adTnt.Rooms != nil && *adTnt.Rooms > 0 {
		second = append(second, fmt.Sprintf("%d комн", *adTnt.Rooms))
	}

	m2 := make([]string, 0, m2Cap)
	if adTnt.M2Main != nil && *adTnt.M2Main > 0 {
		m2 = append(m2, getFloatM2(*adTnt.M2Main))
	}
	if adTnt.M2Living != nil && *adTnt.M2Living > 0 {
		m2 = append(m2, getFloatM2(*adTnt.M2Living))
	}
	if adTnt.M2Kitchen != nil && *adTnt.M2Kitchen > 0 {
		m2 = append(m2, getFloatM2(*adTnt.M2Kitchen))
	}
	if len(m2) > 0 {
		second = append(second, fmt.Sprintf("%s м²", strings.Join(m2, " / ")))
	}

	floor := make([]string, 0, floorCap)
	if adTnt.Floor != nil && *adTnt.Floor > 0 {
		floor = append(floor, fmt.Sprintf("%d", *adTnt.Floor))
	}
	if adTnt.Floors != nil && *adTnt.Floors > 0 {
		floor = append(floor, fmt.Sprintf("%d", *adTnt.Floors))
	}
	if len(floor) > 0 {
		second = append(second, fmt.Sprintf("этаж %s", strings.Join(floor, "/")))
	}

	bathroom := ""
	if adTnt.Bathroom != nil && *adTnt.Bathroom != "" {
		bathroom = fmt.Sprintf("с/у %s", strings.ToLower(*adTnt.Bathroom))
	}

	mess := new(bytes.Buffer)
	if err := t.tmpl.ExecuteTemplate(mess, tmplTgNewAd, map[string]any{
		"URL":      adTnt.URL,
		"Address":  address,
		"Price":    price,
		"M2":       strings.Join(second, "   "),
		"Bathroom": bathroom,
	}); err != nil {
		errs = append(errs, err)
		return errs
	}

	for _, chatID := range tgIDs {
		if len(adTnt.Photos) > 0 {
			err := t.client.SendMediaPhoto(ctx, []string{adTnt.Photos[0]}, []string{mess.String()}, chatID, retryCount)
			if err != nil {
				log.Warnf("notify new ad: media photo ad id %d chat id %d: %s", adTnt.ID, chatID, err)
				err = t.client.SendUploadPhoto(ctx, []string{"static/img/no-image.png"},
					[]string{mess.String()}, chatID, retryCount)
				if err != nil {
					log.Warnf("notify new ad: upload photo ad id %d chat id %d: %s", adTnt.ID, chatID, err)
					err = t.client.SendMessage(ctx, mess.String(), chatID, retryCount)
					if err != nil {
						errs = append(errs, fmt.Errorf("notify new ad: message chat id %d: %w", chatID, err))
					}
				}
			}
		} else {
			err := t.client.SendUploadPhoto(ctx, []string{"static/img/no-image.png"},
				[]string{mess.String()}, chatID, retryCount)
			if err != nil {
				log.Warnf("notify new ad: upload photo ad id %d chat id %d: %s", adTnt.ID, chatID, err)
				err = t.client.SendMessage(ctx, mess.String(), chatID, retryCount)
				if err != nil {
					errs = append(errs, fmt.Errorf("notify new ad: message chat id %d: %w", chatID, err))
				}
			}
		}
	}

	sendAdTotal.WithLabelValues("ok").Add(float64(len(tgIDs) - len(errs)))

	if len(errs) > 0 {
		sendAdTotal.WithLabelValues("error").Add(float64(len(errs)))
		return errs
	}

	return nil
}

func getFloatM2(f float64) string {
	if int(f*roundNumber)%roundNumber > 0 {
		return fmt.Sprintf("%.1f", f)
	}
	return fmt.Sprintf("%.0f", f)
}
