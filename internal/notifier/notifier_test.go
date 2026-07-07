package notifier

import (
	"strings"
	"testing"
	"x-telegram-monitor-bot/internal/domain"
)

func TestFormatTweetMessageIncludesAuthorTextAndURL(t *testing.T) {
	msg := FormatTweetMessage(domain.Tweet{TweetID: "1", XUsername: "openai", Text: "hello", URL: "https://x.com/openai/status/1", Type: "tweet"})
	for _, want := range []string{"Новый твит от @openai", "hello", "https://x.com/openai/status/1"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("message missing %q: %s", want, msg)
		}
	}
}
func TestFormatTweetMessageTruncatesVeryLongText(t *testing.T) {
	msg := FormatTweetMessage(domain.Tweet{XUsername: "a", Text: strings.Repeat("я", 3100), URL: "u"})
	if len([]rune(msg)) > 3100 {
		t.Fatalf("message too long: %d", len([]rune(msg)))
	}
}
