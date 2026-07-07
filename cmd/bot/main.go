package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"x-telegram-monitor-bot/internal/config"
	"x-telegram-monitor-bot/internal/httpserver"
	"x-telegram-monitor-bot/internal/notifier"
	"x-telegram-monitor-bot/internal/storage"
	"x-telegram-monitor-bot/internal/telegram"
	"x-telegram-monitor-bot/internal/twitterapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("config error", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := storage.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connect error", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	twitter := twitterapi.NewClient(cfg.TwitterAPIKey, cfg.TwitterAPIBaseURL, http.DefaultClient)
	telegramClient := telegram.NewClient(cfg.TelegramBotToken, cfg.TelegramAPIBaseURL, http.DefaultClient)
	notify := notifier.New(store, telegramClient, logger)

	bot := telegram.NewBot(store, twitter, telegramClient, logger)
	server := httpserver.New(cfg, store, notify, logger)

	go func() {
		if err := bot.RunLongPolling(ctx, cfg.TelegramPollTimeoutSeconds); err != nil && ctx.Err() == nil {
			logger.Error("telegram polling stopped", "error", err)
			stop()
		}
	}()

	go func() {
		logger.Info("http server starting", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server stopped", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
	logger.Info("shutdown complete")
}
