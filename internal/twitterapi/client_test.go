package twitterapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddUserToMonitorTweetUsesExactEndpointAndField(t *testing.T) {
	var gotPath, gotKey, gotUser string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get("x-api-key")
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotUser = body["x_user_name"]
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success", "msg": "ok"})
	}))
	defer srv.Close()
	c := NewClient("key123", srv.URL, srv.Client())
	if err := c.AddUserToMonitorTweet(context.Background(), "elonmusk"); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/oapi/x_user_stream/add_user_to_monitor_tweet" || gotKey != "key123" || gotUser != "elonmusk" {
		t.Fatalf("got path=%q key=%q user=%q", gotPath, gotKey, gotUser)
	}
}

func TestListMonitoredTweetUsersParsesIDForRemoval(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query_type") != "1" {
			t.Fatalf("query_type=%q", r.URL.Query().Get("query_type"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "msg": "ok", "data": []map[string]any{{"id_for_user": "abc", "x_user_name": "elonmusk", "x_user_screen_name": "elonmusk", "is_monitor_tweet": true}}})
	}))
	defer srv.Close()
	rows, err := NewClient("k", srv.URL, srv.Client()).ListMonitoredTweetUsers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].IDForUser != "abc" {
		t.Fatalf("rows=%+v", rows)
	}
}
