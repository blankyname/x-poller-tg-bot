package xutil

import (
	"regexp"
	"strings"
)

var usernameRe = regexp.MustCompile(`^[A-Za-z0-9_]{1,15}$`)

func NormalizeUsername(input string) (string, bool) {
	s := strings.TrimSpace(input)
	s = strings.TrimPrefix(s, "https://x.com/")
	s = strings.TrimPrefix(s, "https://twitter.com/")
	s = strings.TrimPrefix(s, "http://x.com/")
	s = strings.TrimPrefix(s, "http://twitter.com/")
	s = strings.TrimPrefix(s, "x.com/")
	s = strings.TrimPrefix(s, "twitter.com/")
	s = strings.TrimPrefix(s, "@")
	if idx := strings.IndexAny(s, "/? "); idx >= 0 {
		s = s[:idx]
	}
	if !usernameRe.MatchString(s) {
		return "", false
	}
	return strings.ToLower(s), true
}
