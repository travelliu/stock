// Package tushare is a minimal SDK for the Tushare Pro JSON API.
package tushare

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

const DefaultBaseURL = "https://api.tushare.pro"

// Response is the decoded Tushare envelope. Items is row-major.
type Response struct {
	Fields []string `json:"fields"`
	Items  [][]any  `json:"items"`
}

type envelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// Client is a thin Tushare Pro HTTP client. Construct with NewClient.
type Client struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int
	retryDelay time.Duration
}

// Option configures the client.
type Option func(*Client)

func WithBaseURL(u string) Option        { return func(c *Client) { c.baseURL = u } }
func WithTimeout(d time.Duration) Option { return func(c *Client) { c.httpClient.Timeout = d } }
func WithMaxRetries(n int) Option        { return func(c *Client) { c.maxRetries = n } }
func WithRetryDelay(d time.Duration) Option {
	return func(c *Client) { c.retryDelay = d }
}

// NewClient returns a Tushare client with sensible defaults.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		maxRetries: 2,
		retryDelay: 500 * time.Millisecond,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// BaseURL returns the configured base URL (used in tests).
func (c *Client) BaseURL() string { return c.baseURL }

// Call performs a Tushare API request. `token` is passed per call so each
// caller can use a different token.
func (c *Client) Call(ctx context.Context, token, apiName string, params map[string]any, fields string) (*Response, error) {
	body, err := json.Marshal(map[string]any{
		"api_name": apiName,
		"token":    token,
		"params":   params,
		"fields":   fields,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			jitter := time.Duration(rand.Int63n(int64(c.retryDelay)))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay*time.Duration(attempt) + jitter):
			}
		}
		resp, err := c.doOnce(ctx, body)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isRetryable(err) {
			return nil, err
		}
	}
	return nil, lastErr
}

func (c *Client) doOnce(ctx context.Context, body []byte) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, retryable{err: err}
	}
	defer resp.Body.Close()
	
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, retryable{err: err}
	}
	if resp.StatusCode >= 500 {
		return nil, retryable{err: fmt.Errorf("tushare http %d: %s", resp.StatusCode, string(raw))}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("tushare http %d: %s", resp.StatusCode, string(raw))
	}
	
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("decode envelope: %w; body: %s", err, string(raw))
	}
	if env.Code != 0 {
		return nil, fmt.Errorf("tushare api error %d: %s", env.Code, env.Msg)
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return &Response{}, nil
	}
	var out Response
	if err := json.Unmarshal(env.Data, &out); err != nil {
		return nil, fmt.Errorf("decode data: %w", err)
	}
	return &out, nil
}

type retryable struct{ err error }

func (r retryable) Error() string { return r.err.Error() }
func (r retryable) Unwrap() error { return r.err }

func isRetryable(err error) bool {
	var r retryable
	return errors.As(err, &r)
}
