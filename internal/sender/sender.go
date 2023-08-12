package sender

import (
	"context"
	"fmt"
	"text/template"

	"github.com/sku4/ad-notifier/configs"
	"github.com/sku4/ad-notifier/internal/sender/telegram"
	clientModel "github.com/sku4/ad-parser/pkg/ad/model"
	"github.com/sku4/ad-parser/pkg/ad/street"
)

//go:generate mockgen -source=sender.go -destination=mocks/sender.go

type Telegram interface {
	NotifyNewAd(context.Context, *clientModel.AdTnt, *street.Ext, []int64) []error
}

type Sender struct {
	Telegram
}

func NewSender(ctx context.Context) (*Sender, []error) {
	errs := make([]error, 0)
	cfg := configs.Get(ctx)

	tmpl, err := template.New("ad").ParseGlob(fmt.Sprintf("%s/*.gohtml", cfg.Sender.TemplateFolder))
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}

	tg, err := telegram.NewTelegram(ctx, tmpl)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, errs
	}

	return &Sender{
		Telegram: tg,
	}, nil
}
