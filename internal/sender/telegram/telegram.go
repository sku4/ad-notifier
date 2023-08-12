package telegram

import (
	"context"
	"text/template"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"github.com/sku4/ad-api/pkg/telegram/bot/client"
	"github.com/sku4/ad-notifier/configs"
)

type Telegram struct {
	client client.BotClient
	tmpl   *template.Template
}

func NewTelegram(ctx context.Context, tmpl *template.Template) (*Telegram, error) {
	cfg := configs.Get(ctx)
	tgBot, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, errors.Wrap(err, "init telegram bot")
	}

	tgClient, err := client.NewClient(tgBot)
	if err != nil {
		return nil, errors.Wrap(err, "init telegram client")
	}

	return &Telegram{
		client: tgClient,
		tmpl:   tmpl,
	}, nil
}
