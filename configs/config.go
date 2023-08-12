package configs

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Notifier  `mapstructure:"notifier"`
	Tarantool `mapstructure:"tarantool"`
	Telegram  `mapstructure:"telegram"`
	Sender    `mapstructure:"sender"`
	Rest      `mapstructure:"rest"`
	Features  []string `mapstructure:"features"`
}

type Notifier struct {
	WorkerQueueCount        int `mapstructure:"worker_queue_count"`
	WorkerSubscriptionCount int `mapstructure:"worker_subscription_count"`
}

type Tarantool struct {
	Servers           []string      `mapstructure:"servers"`
	User              string        `mapstructure:"user"`
	Password          string        `mapstructure:"password"`
	Timeout           time.Duration `mapstructure:"timeout"`
	ReconnectInterval time.Duration `mapstructure:"reconnect_interval"`
}

type Telegram struct {
	BotToken string `mapstructure:"bot_token"`
}

type Sender struct {
	TemplateFolder string `mapstructure:"template_folder"`
}

type Rest struct {
	Port           int           `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

func Init() (*Config, error) {
	mainViper := viper.New()
	mainViper.AddConfigPath("configs")
	if err := mainViper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config

	if err := mainViper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading env variables: %w", err)
	}

	cfg.Telegram.BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")

	return &cfg, nil
}

type configKey struct{}

func Set(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, configKey{}, cfg)
}

func Get(ctx context.Context) *Config {
	contextConfig, _ := ctx.Value(configKey{}).(*Config)

	return contextConfig
}
