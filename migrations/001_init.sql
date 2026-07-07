create table if not exists telegram_users (
  id bigserial primary key,
  telegram_id bigint not null unique,
  chat_id bigint not null,
  username text,
  created_at timestamptz not null default now()
);

create table if not exists tracked_accounts (
  id bigserial primary key,
  x_username text not null unique,
  x_user_id text,
  id_for_user text,
  is_active boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists subscriptions (
  id bigserial primary key,
  telegram_user_id bigint not null references telegram_users(id) on delete cascade,
  tracked_account_id bigint not null references tracked_accounts(id) on delete cascade,
  created_at timestamptz not null default now(),
  unique (telegram_user_id, tracked_account_id)
);

create table if not exists tweets (
  tweet_id text primary key,
  x_username text not null,
  text text not null,
  url text not null,
  type text,
  published_at timestamptz,
  raw_json jsonb not null,
  created_at timestamptz not null default now()
);

create table if not exists notifications (
  id bigserial primary key,
  tweet_id text not null references tweets(tweet_id) on delete cascade,
  chat_id bigint not null,
  sent_at timestamptz,
  status text not null,
  error text,
  created_at timestamptz not null default now(),
  unique(tweet_id, chat_id)
);

create index if not exists idx_subscriptions_account on subscriptions(tracked_account_id);
create index if not exists idx_tweets_username on tweets(x_username, published_at desc);
