package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"x-telegram-monitor-bot/internal/config"
	"x-telegram-monitor-bot/internal/domain"
	"x-telegram-monitor-bot/internal/notifier"
	"x-telegram-monitor-bot/internal/storage"
	"x-telegram-monitor-bot/internal/xutil"
)

type Server struct {
	*http.Server
	cfg      config.Config
	store    *storage.Postgres
	notifier *notifier.Notifier
	log      *slog.Logger
}

func New(cfg config.Config, store *storage.Postgres, notifier *notifier.Notifier, log *slog.Logger) *Server {
	s := &Server{cfg: cfg, store: store, notifier: notifier, log: log}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/webhooks/twitterapi", s.twitterapiWebhook)
	s.Server = &http.Server{Addr: cfg.HTTPAddr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	return s
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

type webhookPayload struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	URL       string `json:"url"`
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
	Author    struct {
		UserName   string `json:"userName"`
		Username   string `json:"username"`
		ScreenName string `json:"screen_name"`
	} `json:"author"`
	Tweet *webhookPayload `json:"tweet"`
}

func (s *Server) twitterapiWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.cfg.WebhookSecret != "" && r.Header.Get("X-Webhook-Secret") != s.cfg.WebhookSecret && r.URL.Query().Get("secret") != s.cfg.WebhookSecret {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var raw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	payloads := splitPayload(raw)
	for _, p := range payloads {
		t, ok := payloadToTweet(p, raw)
		if !ok {
			s.log.Warn("unrecognized twitterapi webhook payload")
			continue
		}
		if err := s.notifier.NotifyTweet(r.Context(), t); err != nil {
			s.log.Error("notify tweet failed", "error", err)
			http.Error(w, "notify failed", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func splitPayload(raw json.RawMessage) []webhookPayload {
	var arr []webhookPayload
	if json.Unmarshal(raw, &arr) == nil && len(arr) > 0 {
		return arr
	}
	var p webhookPayload
	_ = json.Unmarshal(raw, &p)
	if p.Tweet != nil {
		return []webhookPayload{*p.Tweet}
	}
	return []webhookPayload{p}
}

func payloadToTweet(p webhookPayload, raw []byte) (domain.Tweet, bool) {
	username := p.Author.UserName
	if username == "" {
		username = p.Author.Username
	}
	if username == "" {
		username = p.Author.ScreenName
	}
	username, ok := xutil.NormalizeUsername(username)
	if !ok || p.ID == "" {
		return domain.Tweet{}, false
	}
	created := time.Now().UTC()
	if p.CreatedAt != "" {
		if tt, err := time.Parse(time.RFC3339, p.CreatedAt); err == nil {
			created = tt
		}
	}
	url := p.URL
	if url == "" {
		url = "https://x.com/" + username + "/status/" + p.ID
	}
	return domain.Tweet{TweetID: p.ID, XUsername: username, Text: strings.TrimSpace(p.Text), URL: url, Type: p.Type, CreatedAtX: created, RawJSON: raw}, true
}
