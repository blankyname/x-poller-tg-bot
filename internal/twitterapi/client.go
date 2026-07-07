package twitterapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	apiKey, baseURL string
	http            *http.Client
}

func NewClient(apiKey, baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{apiKey: apiKey, baseURL: strings.TrimRight(baseURL, "/"), http: httpClient}
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e APIError) Error() string {
	return fmt.Sprintf("twitterapi.io error status=%d message=%s", e.StatusCode, e.Message)
}

type MonitorRow struct {
	IDForUser        string `json:"id_for_user"`
	XUserID          string `json:"x_user_id"`
	XUserName        string `json:"x_user_name"`
	XUserScreenName  string `json:"x_user_screen_name"`
	IsMonitorTweet   bool   `json:"is_monitor_tweet"`
	IsMonitorProfile bool   `json:"is_monitor_profile"`
}

type statusResponse struct {
	Status string       `json:"status"`
	Msg    string       `json:"msg"`
	RuleID string       `json:"rule_id"`
	Data   []MonitorRow `json:"data"`
}

func (c *Client) AddUserToMonitorTweet(ctx context.Context, username string) error {
	var res statusResponse
	if err := c.do(ctx, http.MethodPost, "/oapi/x_user_stream/add_user_to_monitor_tweet", nil, map[string]string{"x_user_name": strings.TrimPrefix(username, "@")}, &res); err != nil {
		return err
	}
	return semanticError(res.Status, res.Msg)
}

func (c *Client) ListMonitoredTweetUsers(ctx context.Context) ([]MonitorRow, error) {
	q := url.Values{"query_type": {"1"}}
	var res statusResponse
	if err := c.do(ctx, http.MethodGet, "/oapi/x_user_stream/get_user_to_monitor_tweet", q, nil, &res); err != nil {
		return nil, err
	}
	if err := semanticError(res.Status, res.Msg); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) RemoveUserFromMonitorTweet(ctx context.Context, idForUser string) error {
	var res statusResponse
	if err := c.do(ctx, http.MethodPost, "/oapi/x_user_stream/remove_user_to_monitor_tweet", nil, map[string]string{"id_for_user": idForUser}, &res); err != nil {
		return err
	}
	return semanticError(res.Status, res.Msg)
}

func (c *Client) do(ctx context.Context, method, path string, q url.Values, body any, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	u := c.baseURL + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return err
	}
	req.Header.Set("x-api-key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return APIError{StatusCode: resp.StatusCode, Message: string(payload)}
	}
	if out != nil && len(payload) > 0 {
		if err := json.Unmarshal(payload, out); err != nil {
			return fmt.Errorf("decode twitterapi response: %w: %s", err, string(payload))
		}
	}
	return nil
}

func semanticError(status, msg string) error {
	if strings.EqualFold(status, "error") {
		return APIError{StatusCode: 200, Message: msg}
	}
	return nil
}
