# X Telegram Monitor Bot

Real-time Telegram bot for X/Twitter account notifications via **twitterapi.io user stream/webhook**.

## Features

- Telegram commands: `/start`, `/add`, `/remove`, `/list`, `/help`.
- One X account can fan out to many Telegram subscribers.
- twitterapi.io real-time monitor endpoint: `/oapi/x_user_stream/add_user_to_monitor_tweet`.
- Webhook receiver: `POST /webhooks/twitterapi`.
- PostgreSQL persistence and tweet deduplication by `tweet_id`.
- Docker Compose deployment.

## Why stream/webhook, not cron polling

For 30+ accounts and low latency, polling `/twitter/user/last_tweets` is expensive and slower. twitterapi.io recommends real-time stream monitoring for 20+ accounts. Steady-state target latency is usually 1–3 seconds end-to-end, depending on provider delivery and Telegram.

## Quick start

```bash
cp .env.example .env
# edit .env: TELEGRAM_BOT_TOKEN, TWITTERAPI_IO_KEY, WEBHOOK_SECRET, PUBLIC_BASE_URL

docker compose up --build
```

Additional docs:

- [`DEPLOYMENT.md`](./DEPLOYMENT.md) — local/Docker deployment and webhook setup.
- [`SECURITY.md`](./SECURITY.md) — secret handling and leak checks.
- [`AGENTS.md`](./AGENTS.md) — instructions for future AI agents working in this repo.

Webhook URL to configure in twitterapi.io/dashboard:

```text
https://your-domain.example/webhooks/twitterapi?secret=WEBHOOK_SECRET
```

Alternatively send the secret as header:

```text
X-Webhook-Secret: WEBHOOK_SECRET
```

## Commands

```text
/start
/add openai
/remove openai
/list
/help
```

## Local DB migration without Docker

```bash
export DATABASE_URL='postgres://xmonitor:xmonitor@localhost:5432/xmonitor?sslmode=disable'
go run ./cmd/migrate ./migrations
```

## Tests

```bash
go test ./...
```

## Notes for twitterapi.io

Important exact fields:

- Add monitor: body `{ "x_user_name": "elonmusk" }`
- List monitors: `GET /oapi/x_user_stream/get_user_to_monitor_tweet?query_type=1`
- Remove monitor: body `{ "id_for_user": "..." }`

Do not hardcode API keys or commit `.env`.
