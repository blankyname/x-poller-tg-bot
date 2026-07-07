package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"x-telegram-monitor-bot/internal/storage"
	"x-telegram-monitor-bot/internal/twitterapi"
	"x-telegram-monitor-bot/internal/xutil"
)

type Bot struct {
	store   *storage.Postgres
	twitter *twitterapi.Client
	tg      *Client
	log     *slog.Logger
}

func NewBot(store *storage.Postgres, twitter *twitterapi.Client, tg *Client, log *slog.Logger) *Bot {
	return &Bot{store: store, twitter: twitter, tg: tg, log: log}
}

func (b *Bot) RunLongPolling(ctx context.Context, timeoutSeconds int) error {
	var offset int64
	for ctx.Err() == nil {
		updates, err := b.tg.GetUpdates(ctx, offset, timeoutSeconds)
		if err != nil {
			b.log.Warn("getUpdates failed", "error", err)
			continue
		}
		for _, u := range updates {
			if u.UpdateID >= offset {
				offset = u.UpdateID + 1
			}
			if u.Message != nil {
				b.handleMessage(ctx, *u.Message)
			}
		}
	}
	return ctx.Err()
}

func (b *Bot) handleMessage(ctx context.Context, msg Message) {
	userID := msg.Chat.ID
	username := msg.Chat.Username
	if msg.From != nil {
		userID = msg.From.ID
		username = msg.From.Username
	}
	if err := b.store.UpsertTelegramUser(ctx, userID, msg.Chat.ID, username); err != nil {
		b.reply(ctx, msg.Chat.ID, "Ошибка регистрации пользователя.")
		return
	}
	parts := strings.Fields(strings.TrimSpace(msg.Text))
	if len(parts) == 0 {
		return
	}
	cmd := strings.Split(parts[0], "@")[0]
	switch cmd {
	case "/start", "/help":
		b.reply(ctx, msg.Chat.ID, helpText())
	case "/add":
		if len(parts) < 2 {
			b.reply(ctx, msg.Chat.ID, "Использование: /add username")
			return
		}
		xname, ok := xutil.NormalizeUsername(parts[1])
		if !ok {
			b.reply(ctx, msg.Chat.ID, "Некорректный X username.")
			return
		}
		if err := b.twitter.AddUserToMonitorTweet(ctx, xname); err != nil {
			b.log.Warn("twitter monitor add failed", "username", xname, "error", err)
		}
		if err := b.store.Subscribe(ctx, userID, msg.Chat.ID, xname); err != nil {
			b.reply(ctx, msg.Chat.ID, "Ошибка сохранения подписки.")
			return
		}
		b.reply(ctx, msg.Chat.ID, fmt.Sprintf("Подписка добавлена: @%s", xname))
	case "/remove":
		if len(parts) < 2 {
			b.reply(ctx, msg.Chat.ID, "Использование: /remove username")
			return
		}
		xname, ok := xutil.NormalizeUsername(parts[1])
		if !ok {
			b.reply(ctx, msg.Chat.ID, "Некорректный X username.")
			return
		}
		if err := b.store.Unsubscribe(ctx, userID, xname); err != nil {
			b.reply(ctx, msg.Chat.ID, "Ошибка удаления подписки.")
			return
		}
		b.reply(ctx, msg.Chat.ID, fmt.Sprintf("Подписка удалена: @%s", xname))
	case "/list":
		items, err := b.store.ListSubscriptions(ctx, userID)
		if err != nil {
			b.reply(ctx, msg.Chat.ID, "Ошибка чтения подписок.")
			return
		}
		if len(items) == 0 {
			b.reply(ctx, msg.Chat.ID, "Список пуст. Добавь аккаунт: /add username")
			return
		}
		b.reply(ctx, msg.Chat.ID, "Отслеживаются:\n@"+strings.Join(items, "\n@"))
	default:
		b.reply(ctx, msg.Chat.ID, "Неизвестная команда. /help")
	}
}

func (b *Bot) reply(ctx context.Context, chatID int64, text string) {
	if err := b.tg.SendMessage(ctx, chatID, text); err != nil {
		b.log.Warn("telegram reply failed", "error", err)
	}
}
func helpText() string {
	return "Команды:\n/start — регистрация\n/add username — подписаться на X аккаунт\n/remove username — отписаться\n/list — список подписок\n/help — помощь"
}
