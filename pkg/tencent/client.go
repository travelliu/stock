// Package tencent fetches intraday quotes from Tencent Finance (qt.gtimg.cn).
package tencent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"stock/pkg/utils"
	"strconv"
	"strings"
	"time"

	"stock/pkg/models"
)

const (
	defaultBaseURL = "https://qt.gtimg.cn/q="

	idxName      = 1
	idxPrice     = 3
	idxPrevClose = 4
	idxOpen      = 5
	idxVol       = 6
	idxQuoteTime = 30
	idxChange    = 31
	idxChangePct = 32
	idxHigh      = 33
	idxLow       = 34
	idxAmount    = 37
	idxLimitUp   = 47
	idxLimitDown = 48
	idxMinFields = 49
)

// Client fetches realtime quotes from Tencent Finance.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Option configures Client.
type Option func(*Client)

// WithBaseURL overrides the base URL (used in tests to point at httptest.Server).
func WithBaseURL(u string) Option { return func(c *Client) { c.baseURL = u } }

// NewClient returns a Client with sensible defaults.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// FetchQuotes retrieves realtime quotes for the given tsCode list (e.g. "600519.SH").
// Returns partial results on per-stock parse errors.
func (c *Client) FetchQuotes(ctx context.Context, tsCodes []string) ([]*models.RealtimeQuote, error) {
	if len(tsCodes) == 0 {
		return nil, nil
	}
	url := c.baseURL + strings.Join(tsToCodes(tsCodes), ",")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Referer", "https://finance.qq.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch quotes: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return parseQuotes(body), nil
}

func parseQuotes(body []byte) []*models.RealtimeQuote {
	now := time.Now()
	var out []*models.RealtimeQuote
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "v_") {
			continue
		}
		eqIdx := strings.Index(line, "=")
		if eqIdx < 0 {
			continue
		}
		tencentCode := line[2:eqIdx] // e.g. "sh600519"

		start := strings.Index(line, `"`)
		end := strings.LastIndex(line, `"`)
		if start < 0 || end <= start {
			continue
		}
		content := line[start+1 : end]
		fields := strings.Split(content, "~")
		if len(fields) < idxMinFields {
			continue
		}
		out = append(out, &models.RealtimeQuote{
			TsCode:    tencentToTs(tencentCode),
			Price:     parseFloat(fields[idxPrice]),
			PrevClose: parseFloat(fields[idxPrevClose]),
			Open:      parseFloat(fields[idxOpen]),
			Vol:       parseFloat(fields[idxVol]),
			High:      parseFloat(fields[idxHigh]),
			Low:       parseFloat(fields[idxLow]),
			Amount:    parseFloat(fields[idxAmount]),
			Change:    parseFloat(fields[idxChange]),
			ChangePct: parseFloat(fields[idxChangePct]),
			LimitUp:   parseFloat(fields[idxLimitUp]),
			LimitDown: parseFloat(fields[idxLimitDown]),
			QuoteTime: fields[idxQuoteTime],
			UpdatedAt: now,
		})
	}
	return out
}

// tsToCodes converts ["600519.SH","000858.SZ"] -> ["sh600519","sz000858"].
func tsToCodes(tsCodes []string) []string {
	out := make([]string, 0, len(tsCodes))
	for _, ts := range tsCodes {
		out = append(out, utils.ToTencentCode(ts))
	}
	return out
}

// tencentToTs converts "sh600519" -> "600519.SH".
func tencentToTs(code string) string {
	if len(code) < 3 {
		return code
	}
	num := code[2:]
	return num
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}
