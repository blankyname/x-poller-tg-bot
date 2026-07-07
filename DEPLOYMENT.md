# Deployment Guide

## Local development

```bash
cp .env.example .env
# edit .env

go test ./...
go run ./cmd/migrate ./migrations
go run ./cmd/bot
```

## Docker Compose

```bash
cp .env.example .env
# edit .env

docker compose up --build
```

Services:

- `db` — PostgreSQL 16.
- `migrate` — applies SQL migrations.
- `bot` — Telegram polling + HTTP webhook server.

## Public webhook

twitterapi.io needs a public HTTPS URL:

```text
https://YOUR_DOMAIN/webhooks/twitterapi?secret=WEBHOOK_SECRET
```

or header:

```text
X-Webhook-Secret: WEBHOOK_SECRET
```

For local testing, use a tunnel such as cloudflared/ngrok and set `PUBLIC_BASE_URL` accordingly.

## Telegram commands

```text
/start
/add openai
/remove openai
/list
/help
```

## twitterapi.io stream setup

The bot calls the user monitoring endpoint when `/add` is used:

```http
POST /oapi/x_user_stream/add_user_to_monitor_tweet
{ "x_user_name": "openai" }
```

You still need an active twitterapi.io monitoring/stream subscription and webhook delivery configured in the provider dashboard.

## Operational checks

- `GET /healthz` returns `ok`.
- Bot logs are JSON on stdout.
- Duplicate tweet events should create only one row in `tweets`.
- Failed Telegram sends are recorded in `notifications` with status `failed`.
