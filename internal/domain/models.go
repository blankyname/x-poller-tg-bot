package domain

import "time"

type TelegramUser struct {
	ID         int64
	TelegramID int64
	ChatID     int64
	Username   string
	CreatedAt  time.Time
}

type TrackedAccount struct {
	ID        int64
	XUsername string
	XUserID   string
	IDForUser string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Tweet struct {
	TweetID    string
	XUsername  string
	Text       string
	URL        string
	Type       string
	CreatedAtX time.Time
	RawJSON    []byte
}
