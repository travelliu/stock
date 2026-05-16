// Package baidu fetches concept blocks and fund flow from Baidu PAE (finance.pae.baidu.com).
package baidu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://finance.pae.baidu.com"

var defaultHeaders = map[string]string{
	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/117.0.0.0 Safari/537.36",
	"Accept":     "application/vnd.finance-web.v1+json",
	"Origin":     "https://gushitong.baidu.com",
	"Referer":    "https://gushitong.baidu.com/",
}

// Client calls Baidu PAE finance APIs.
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

// BlockItem is a single board entry (industry, concept, or region).
type BlockItem struct {
	Name      string `json:"name"`
	ChangePct string `json:"change_pct"`
}

// ConceptBlocks groups a stock's board memberships by type.
type ConceptBlocks struct {
	Industry    []BlockItem `json:"industry"`
	Concept     []BlockItem `json:"concept"`
	Region      []BlockItem `json:"region"`
	ConceptTags []string    `json:"concept_tags"`
}

// FundFlowDay is one day's fund flow summary for a stock.
type FundFlowDay struct {
	Date        string `json:"date"`
	Close       string `json:"close"`
	ChangePct   string `json:"change_pct"`
	SuperNetIn  string `json:"super_net_in"`  // 超大单净流入 (万元)
	LargeNetIn  string `json:"large_net_in"`  // 大单净流入
	MediumNetIn string `json:"medium_net_in"` // 中单
	LittleNetIn string `json:"little_net_in"` // 小单
	MainIn      string `json:"main_in"`       // 主力净流入
}

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("baidu: build request: %w", err)
	}
	for k, v := range defaultHeaders {
		req.Header.Set(k, v)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("baidu: request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("baidu: read body: %w", err)
	}
	return body, nil
}

// FetchConceptBlocks returns industry/concept/region board memberships for a stock.
// code is a 6-digit stock code (no exchange suffix).
func (c *Client) FetchConceptBlocks(ctx context.Context, code string) (*ConceptBlocks, error) {
	path := fmt.Sprintf("/api/getrelatedblock?code=%s&market=ab&typeCode=all&finClientType=pc", code)
	body, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var raw struct {
		ResultCode any               `json:"ResultCode"`
		Result     []json.RawMessage `json:"Result"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("baidu: parse blocks: %w", err)
	}
	if fmt.Sprintf("%v", raw.ResultCode) != "0" {
		return nil, fmt.Errorf("baidu: API error ResultCode=%v", raw.ResultCode)
	}

	out := &ConceptBlocks{}
	for _, rawBlock := range raw.Result {
		var block struct {
			Type string `json:"type"`
			List []struct {
				Name     string `json:"name"`
				Increase string `json:"increase"`
			} `json:"list"`
		}
		if err := json.Unmarshal(rawBlock, &block); err != nil {
			continue
		}
		for _, item := range block.List {
			entry := BlockItem{Name: item.Name, ChangePct: item.Increase}
			switch {
			case strings.Contains(block.Type, "行业"):
				out.Industry = append(out.Industry, entry)
			case strings.Contains(block.Type, "概念"):
				out.Concept = append(out.Concept, entry)
				out.ConceptTags = append(out.ConceptTags, item.Name)
			case strings.Contains(block.Type, "地域"):
				out.Region = append(out.Region, entry)
			}
		}
	}
	return out, nil
}

// FetchFundFlowHistory returns daily fund flow for the last `days` trading days.
func (c *Client) FetchFundFlowHistory(ctx context.Context, code string, days int) ([]FundFlowDay, error) {
	path := fmt.Sprintf("/vapi/v1/fundsortlist?code=%s&market=ab&pn=0&rn=%d&finClientType=pc", code, days)
	body, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var raw struct {
		ResultCode any `json:"ResultCode"`
		Result     struct {
			List []map[string]any `json:"list"`
		} `json:"Result"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("baidu: parse fund flow history: %w", err)
	}
	if fmt.Sprintf("%v", raw.ResultCode) != "0" {
		return nil, fmt.Errorf("baidu: API error ResultCode=%v", raw.ResultCode)
	}

	out := make([]FundFlowDay, 0, len(raw.Result.List))
	for _, item := range raw.Result.List {
		out = append(out, FundFlowDay{
			Date:        str(item["showtime"]),
			Close:       str(item["closepx"]),
			ChangePct:   str(item["ratio"]),
			SuperNetIn:  str(item["superNetIn"]),
			LargeNetIn:  str(item["largeNetIn"]),
			MediumNetIn: str(item["mediumNetIn"]),
			LittleNetIn: str(item["littleNetIn"]),
			MainIn:      str(item["extMainIn"]),
		})
	}
	return out, nil
}

func str(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
