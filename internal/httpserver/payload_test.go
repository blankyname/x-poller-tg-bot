package httpserver

import (
	"encoding/json"
	"testing"
)

func TestPayloadToTweetBuildsURLAndNormalizesAuthor(t *testing.T) {
	raw := json.RawMessage(`{"id":"123","text":"hi","author":{"userName":"OpenAI"}}`)
	parts := splitPayload(raw)
	tw, ok := payloadToTweet(parts[0], raw)
	if !ok {
		t.Fatal("payload not parsed")
	}
	if tw.XUsername != "openai" || tw.URL != "https://x.com/openai/status/123" || tw.Text != "hi" {
		t.Fatalf("tweet=%+v", tw)
	}
}
