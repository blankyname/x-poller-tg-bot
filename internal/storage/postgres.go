package storage

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"x-telegram-monitor-bot/internal/domain"
)

type Postgres struct{ pool *pgxpool.Pool }

func NewPostgres(ctx context.Context, databaseURL string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Postgres{pool: pool}, nil
}
func (p *Postgres) Close() { p.pool.Close() }

func (p *Postgres) UpsertTelegramUser(ctx context.Context, telegramID, chatID int64, username string) error {
	_, err := p.pool.Exec(ctx, `insert into telegram_users (telegram_id, chat_id, username) values ($1,$2,$3) on conflict (telegram_id) do update set chat_id=excluded.chat_id, username=excluded.username`, telegramID, chatID, username)
	return err
}

func (p *Postgres) Subscribe(ctx context.Context, telegramID, chatID int64, xUsername string) error {
	xUsername = strings.ToLower(strings.TrimPrefix(xUsername, "@"))
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var userID, accountID int64
	if err := tx.QueryRow(ctx, `insert into telegram_users (telegram_id, chat_id) values ($1,$2) on conflict (telegram_id) do update set chat_id=excluded.chat_id returning id`, telegramID, chatID).Scan(&userID); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, `insert into tracked_accounts (x_username) values ($1) on conflict (x_username) do update set is_active=true, updated_at=now() returning id`, xUsername).Scan(&accountID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `insert into subscriptions (telegram_user_id, tracked_account_id) values ($1,$2) on conflict do nothing`, userID, accountID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *Postgres) Unsubscribe(ctx context.Context, telegramID int64, xUsername string) error {
	_, err := p.pool.Exec(ctx, `delete from subscriptions s using telegram_users u, tracked_accounts a where s.telegram_user_id=u.id and s.tracked_account_id=a.id and u.telegram_id=$1 and a.x_username=$2`, telegramID, strings.ToLower(strings.TrimPrefix(xUsername, "@")))
	return err
}

func (p *Postgres) ListSubscriptions(ctx context.Context, telegramID int64) ([]string, error) {
	rows, err := p.pool.Query(ctx, `select a.x_username from subscriptions s join telegram_users u on u.id=s.telegram_user_id join tracked_accounts a on a.id=s.tracked_account_id where u.telegram_id=$1 order by a.x_username`, telegramID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (p *Postgres) SubscribersForAccount(ctx context.Context, xUsername string) ([]int64, error) {
	rows, err := p.pool.Query(ctx, `select distinct u.chat_id from subscriptions s join telegram_users u on u.id=s.telegram_user_id join tracked_accounts a on a.id=s.tracked_account_id where a.x_username=$1`, strings.ToLower(strings.TrimPrefix(xUsername, "@")))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (p *Postgres) SaveTweetIfNew(ctx context.Context, t domain.Tweet) (bool, error) {
	if t.RawJSON == nil {
		t.RawJSON, _ = json.Marshal(t)
	}
	tag, err := p.pool.Exec(ctx, `insert into tweets (tweet_id, x_username, text, url, type, published_at, raw_json) values ($1,$2,$3,$4,$5,$6,$7) on conflict (tweet_id) do nothing`, t.TweetID, strings.ToLower(t.XUsername), t.Text, t.URL, t.Type, t.CreatedAtX, t.RawJSON)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (p *Postgres) MarkNotification(ctx context.Context, tweetID string, chatID int64, status, errMsg string) error {
	_, err := p.pool.Exec(ctx, `insert into notifications (tweet_id, chat_id, status, error) values ($1,$2,$3,$4) on conflict (tweet_id, chat_id) do update set status=excluded.status, error=excluded.error, sent_at=case when excluded.status='sent' then now() else notifications.sent_at end`, tweetID, chatID, status, errMsg)
	return err
}
