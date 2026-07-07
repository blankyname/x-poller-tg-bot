package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"x-telegram-monitor-bot/internal/domain"
	"x-telegram-monitor-bot/internal/storage"
	"x-telegram-monitor-bot/internal/telegram"
)

type Notifier struct {
	store *storage.Postgres
	tg    *telegram.Client
	log   *slog.Logger
}

func New(store *storage.Postgres, tg *telegram.Client, log *slog.Logger) *Notifier {
	return &Notifier{store: store, tg: tg, log: log}
}

func FormatTweetMessage(t domain.Tweet) string {
	kind := "Новый твит"
	if t.Type != "" && t.Type != "tweet" {
		kind = "Новое событие X (" + t.Type + ")"
	}
	text := strings.TrimSpace(t.Text)
	if text == "" {
		text = "[без текста]"
	}
	if len([]rune(text)) > 3000 {
		r := []rune(text)
		text = string(r[:3000]) + "…"
	}
	return fmt.Sprintf("%s от @%s\n\n%s\n\n%s", kind, t.XUsername, text, t.URL)
}

func (n *Notifier) NotifyTweet(ctx context.Context, t domain.Tweet) error {
	isNew, err := n.store.SaveTweetIfNew(ctx, t)
	if err != nil {
		return err
	}
	if !isNew {
		return nil
	}
	chatIDs, err := n.store.SubscribersForAccount(ctx, t.XUsername)
	if err != nil {
		return err
	}
	msg := FormatTweetMessage(t)
	for _, chatID := range chatIDs {
		if err := n.tg.SendMessage(ctx, chatID, msg); err != nil {
			n.log.Warn("send notification failed", "chat_id", chatID, "tweet_id", t.TweetID, "error", err)
			_ = n.store.MarkNotification(ctx, t.TweetID, chatID, "failed", err.Error())
			continue
		}
		_ = n.store.MarkNotification(ctx, t.TweetID, chatID, "sent", "")
	}
	return nil
}
