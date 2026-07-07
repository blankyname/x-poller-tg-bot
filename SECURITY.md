# Security Notes

## Current status

This repository is designed to contain **no live credentials**. Runtime secrets must be supplied through environment variables or deployment secrets.

Allowed placeholders:

- `TELEGRAM_BOT_TOKEN=replace_with_botfather_token`
- `TWITTERAPI_IO_KEY=replace_with_twitterapi_io_key`
- `WEBHOOK_SECRET=replace_with_random_webhook_secret`

## Required runtime secrets

Set these outside git, for example in `.env`, Docker secrets, or your hosting provider's secret manager:

```bash
TELEGRAM_BOT_TOKEN=...
TWITTERAPI_IO_KEY=...
WEBHOOK_SECRET=...
DATABASE_URL=...
```

`.env` is ignored by git.

## If a token was exposed

If any GitHub, Telegram, twitterapi.io, or X credential was pasted into chat, logs, terminal history, screenshots, or committed by mistake:

1. Revoke it at the provider dashboard immediately.
2. Generate a fresh token.
3. Update deployment secrets.
4. Audit git history before making the repository public.

## Secret scan commands

Run before push:

```bash
git grep -n -I -E '(ghp_|github_pat_|x-access-token|TELEGRAM_BOT_TOKEN=[^[:space:]]{12,}|TWITTERAPI_IO_KEY=[^[:space:]]{12,}|WEBHOOK_SECRET=[^[:space:]]{12,}|login_cookies|totp_secret)' -- . || true
git remote -v
git config --local --list
```

The remote URL must look like:

```text
https://github.com/OWNER/REPO.git
```

It must **not** contain `x-access-token`, `ghp_`, or any other credential.
