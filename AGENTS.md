# Agent Operating Instructions

This repository contains a Go Telegram bot for real-time X/Twitter account monitoring via twitterapi.io.

## Mission

Maintain a low-latency, production-oriented service that:

- lets Telegram users subscribe to X/Twitter accounts;
- registers monitored accounts with twitterapi.io user stream endpoints;
- receives tweet events through `POST /webhooks/twitterapi`;
- deduplicates by `tweet_id`;
- fans out notifications to subscribed Telegram chats.

## Architecture

Key directories:

- `cmd/bot/` — main service entrypoint.
- `cmd/migrate/` — SQL migration runner.
- `internal/config/` — environment config loading.
- `internal/twitterapi/` — twitterapi.io REST client.
- `internal/telegram/` — Telegram Bot API client and command loop.
- `internal/httpserver/` — health endpoint and twitterapi.io webhook receiver.
- `internal/notifier/` — message formatting and fan-out delivery.
- `internal/storage/` — PostgreSQL persistence.
- `internal/xutil/` — X username normalization helpers.
- `migrations/` — PostgreSQL schema.

## Required checks before declaring work done

Run from the repository root:

```bash
go test ./...
go build ./cmd/bot ./cmd/migrate
```

If Go is not globally installed on this Windows host, Hermes previously used:

```bash
export PATH="$HOME/AppData/Local/hermes/tools/go/bin:$PATH"
```

## Security rules

Never commit real secrets. The only allowed secret-looking values are placeholders in `.env.example`.

Forbidden in tracked files:

- real Telegram bot tokens;
- real `TWITTERAPI_IO_KEY` values;
- GitHub tokens (`ghp_...`, `github_pat_...`);
- real `WEBHOOK_SECRET` values;
- database passwords beyond local Docker defaults;
- `login_cookies`, proxies, or X account credentials.

Before committing or pushing, run at least:

```bash
git grep -n -I -E '(ghp_|github_pat_|x-access-token|TELEGRAM_BOT_TOKEN=[^[:space:]]{12,}|TWITTERAPI_IO_KEY=[^[:space:]]{12,}|WEBHOOK_SECRET=[^[:space:]]{12,})' -- . || true
git remote -v
git config --local --list
```

Expected findings may include only placeholders in `.env.example`; `git remote -v` must not contain embedded credentials.

## twitterapi.io specifics

Use exact field names from the installed `twitterapi-io` skill / docs:

- Add monitor: `POST /oapi/x_user_stream/add_user_to_monitor_tweet` with body `{ "x_user_name": "handle" }`.
- List monitors: `GET /oapi/x_user_stream/get_user_to_monitor_tweet?query_type=1`.
- Remove monitor: `POST /oapi/x_user_stream/remove_user_to_monitor_tweet` with body `{ "id_for_user": "..." }`.

Do not build frequent polling around `/twitter/user/last_tweets` for real-time delivery.

## Implementation conventions

- Keep API keys in env vars only.
- Keep SQL parameterized through pgx; never format user input into SQL strings.
- Keep Telegram messages concise and safe for plain Markdown/text rendering.
- Treat webhook payload shapes defensively; twitterapi.io responses can vary.
- Deduplicate before notification fan-out.
- Prefer small, tested changes.

## Human handoff notes

If asked to deploy, first confirm:

1. Telegram bot token from BotFather.
2. twitterapi.io key and active stream subscription.
3. Public HTTPS endpoint for `/webhooks/twitterapi`.
4. Chosen deployment target: VPS Docker Compose, cloud service, or tunnel for local testing.

Do not ask for secrets unless they are required for the current action.
