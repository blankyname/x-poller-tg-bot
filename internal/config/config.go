package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr                   string
	PublicBaseURL              string
	WebhookSecret              string
	DatabaseURL                string
	TelegramBotToken           string
	TelegramAPIBaseURL         string
	TelegramPollTimeoutSeconds int
	TwitterAPIKey              string
	TwitterAPIBaseURL          string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:                   env("HTTP_ADDR", ":8080"),
		PublicBaseURL:              strings.TrimRight(env("PUBLIC_BASE_URL", ""), "/"),
		WebhookSecret:              env("WEBHOOK_SECRET", ""),
		DatabaseURL:                env("DATABASE_URL", ""),
		TelegramBotToken:           env("TELEGRAM_BOT_TOKEN", ""),
		TelegramAPIBaseURL:         strings.TrimRight(env("TELEGRAM_API_BASE_URL", "https://api.telegram.org"), "/"),
		TwitterAPIKey:              env("TWITTERAPI_IO_KEY", ""),
		TwitterAPIBaseURL:          strings.TrimRight(env("TWITTERAPI_IO_BASE_URL", "https://api.twitterapi.io"), "/"),
		TelegramPollTimeoutSeconds: envInt("TELEGRAM_POLL_TIMEOUT_SECONDS", 25),
	}
	var missing []string
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.TelegramBotToken == "" {
		missing = append(missing, "TELEGRAM_BOT_TOKEN")
	}
	if cfg.TwitterAPIKey == "" {
		missing = append(missing, "TWITTERAPI_IO_KEY")
	}
	if len(missing) > 0 {
		return cfg, fmt.Errorf("missing required env: %s", strings.Join(missing, ", "))
	}
	if cfg.WebhookSecret == "" {
		return cfg, errors.New("WEBHOOK_SECRET is required to protect /webhooks/twitterapi")
	}
	return cfg, nil
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func envInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
