package xutil

import "testing"

func TestNormalizeUsernameAcceptsHandlesAndUrls(t *testing.T) {
	cases := map[string]string{" @ElonMusk ": "elonmusk", "https://x.com/OpenAI/status/1": "openai", "twitter.com/sama": "sama"}
	for in, want := range cases {
		got, ok := NormalizeUsername(in)
		if !ok || got != want {
			t.Fatalf("NormalizeUsername(%q)=%q,%v want %q,true", in, got, ok, want)
		}
	}
}

func TestNormalizeUsernameRejectsInvalid(t *testing.T) {
	for _, in := range []string{"", "name-with-dash", "abcdefghijklmnop", "bad!"} {
		if got, ok := NormalizeUsername(in); ok {
			t.Fatalf("NormalizeUsername(%q)=%q,true want false", in, got)
		}
	}
}
