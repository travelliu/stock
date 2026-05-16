// Package baidu fetches concept blocks and fund flow from Baidu PAE (finance.pae.baidu.com).
package baidu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	Describe  string `json:"describe,omitempty"` // e.g. "申万一级"
}

// ConceptBlocks groups a stock's board memberships by type.
type ConceptBlocks struct {
	Industry    []BlockItem `json:"industry"`
	Concept     []BlockItem `json:"concept"`
	Region      []BlockItem `json:"region"`
	ConceptTags []string    `json:"concept_tags"`
}

// OrderFlowGroup is fund flow breakdown for one order-size category.
type OrderFlowGroup struct {
	NetTurnover     string `json:"net_turnover"` // 净流入 (亿元)
	TurnoverIn      string `json:"turnover_in"`
	TurnoverOut     string `json:"turnover_out"`
	TurnoverInRate  string `json:"turnover_in_rate"`
	TurnoverOutRate string `json:"turnover_out_rate"`
}

// OrderFlowGroups is the shared order-size breakdown embedded in both FundFlowLevel and StockFundFlow.
type OrderFlowGroups struct {
	Super  OrderFlowGroup `json:"super"`
	Large  OrderFlowGroup `json:"large"`
	Medium OrderFlowGroup `json:"medium"`
	Little OrderFlowGroup `json:"little"`
}

// MainFlow is the headline main-force flow numbers.
type MainFlow struct {
	MainIn    string `json:"main_in"`
	MainOut   string `json:"main_out"`
	MainNetIn string `json:"main_net_in"`
}

// RecentAgg is an aggregate net-flow over a recent window.
type RecentAgg struct {
	Key   string `json:"key"`   // "近三日", "近五日", "近十日", "近二十日"
	Value string `json:"value"` // net flow in unit below
}

// IndustryInfo identifies the sector a stock belongs to.
type IndustryInfo struct {
	Name string `json:"name"`
	Desc string `json:"desc"` // "申万一级" or "申万二级"
}

// FundFlowLevel is the fund flow data for one industry-classification level.
type FundFlowLevel struct {
	OrderFlowGroups
	Belongs    string       `json:"belongs"` // "stocklevelone" / "stockleveltwo"
	UpdateTime string       `json:"update_time"`
	Unit       string       `json:"unit"` // "亿"
	Industry   IndustryInfo `json:"industry"`
	TodayMain  MainFlow     `json:"today_main"`
	Recently   []RecentAgg  `json:"recently"`
}

// StockFundFlow is today's per-stock fund flow breakdown (个股资金分布).
type StockFundFlow struct {
	OrderFlowGroups
	TodayMain        MainFlow `json:"today_main"`
	TurnoverInTotal  string   `json:"turnover_in_total"`
	TurnoverOutTotal string   `json:"turnover_out_total"`
	TurnoverNetTotal string   `json:"turnover_net_total"`
	MainNetTotal     string   `json:"main_net_total"`
	Unit             string   `json:"unit"`
	UpdateTime       string   `json:"update_time"`
	StockStatus      string   `json:"stock_status"`
}

// FundFlow holds the fund flow data for a stock: per-stock breakdown and industry levels.
type FundFlow struct {
	StockFlow *StockFundFlow   `json:"stock_flow"`
	Levels    []*FundFlowLevel `json:"levels"`
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
	stockParam := fmt.Sprintf(`{"market":"ab","type":"stock","code":"%s"}`, code)
	path := "/api/getrelatedblock?stock=" + url.QueryEscape(stockParam) + "&finClientType=pc"
	body, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var raw struct {
		ResultCode any `json:"ResultCode"`
		Result     map[string][]struct {
			Name string `json:"name"`
			List []struct {
				Name     string `json:"name"`
				Ratio    string `json:"ratio"`
				Describe string `json:"describe"`
			} `json:"list"`
		} `json:"Result"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("baidu: parse blocks: %w", err)
	}
	if fmt.Sprintf("%v", raw.ResultCode) != "0" {
		return nil, fmt.Errorf("baidu: API error ResultCode=%v", raw.ResultCode)
	}

	out := &ConceptBlocks{}
	for _, group := range raw.Result[code] {
		for _, item := range group.List {
			entry := BlockItem{Name: item.Name, ChangePct: item.Ratio, Describe: item.Describe}
			switch group.Name {
			case "行业":
				out.Industry = append(out.Industry, entry)
			case "概念":
				out.Concept = append(out.Concept, entry)
				out.ConceptTags = append(out.ConceptTags, item.Name)
			case "地域":
				out.Region = append(out.Region, entry)
			}
		}
	}
	return out, nil
}

// FetchFundFlow returns today's fund flow breakdown by industry level (申万一级/二级).
// URL: /vapi/v1/fundflow?finance_type=stock&fund_flow_type=&market=ab&code={code}&type=stock&finClientType=pc
func (c *Client) FetchFundFlow(ctx context.Context, code string) (*FundFlow, error) {
	path := fmt.Sprintf(
		"/vapi/v1/fundflow?finance_type=stock&fund_flow_type=&market=ab&code=%s&type=stock&finClientType=pc",
		code,
	)
	body, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var raw struct {
		ResultCode any `json:"ResultCode"`
		Result     struct {
			Content struct {
				FundFlowBlock struct {
					Result []struct {
						Belongs       string      `json:"belongs"`
						UpdateTime    string      `json:"updateTime"`
						Unit          string      `json:"unit"`
						Industry      rawIndustry `json:"industry"`
						TodayMainFlow rawMainFlow `json:"todayMainFlow"`
						SuperGrp      rawGroup    `json:"superGrp"`
						LargeGrp      rawGroup    `json:"largeGrp"`
						MediumGrp     rawGroup    `json:"mediumGrp"`
						LittleGrp     rawGroup    `json:"littleGrp"`
						Recently      []RecentAgg `json:"recently"`
					} `json:"result"`
				} `json:"fundFlowBlock"`
				FundFlowSpread struct {
					Result struct {
						SuperGrp         rawGroup    `json:"superGrp"`
						LargeGrp         rawGroup    `json:"largeGrp"`
						MediumGrp        rawGroup    `json:"mediumGrp"`
						LittleGrp        rawGroup    `json:"littleGrp"`
						TodayMainFlow    rawMainFlow `json:"todayMainFlow"`
						TurnoverInTotal  string      `json:"turnoverInTotal"`
						TurnoverOutTotal string      `json:"turnoverOutTotal"`
						TurnoverNetTotal string      `json:"turnoverNetTotal"`
						MainNetTotal     string      `json:"mainNetTotal"`
						Unit             string      `json:"unit"`
						UpdateTime       string      `json:"updateTime"`
						StockStatus      string      `json:"stockStatus"`
					} `json:"result"`
				} `json:"fundFlowSpread"`
			} `json:"content"`
		} `json:"Result"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("baidu: parse fund flow: %w", err)
	}
	if fmt.Sprintf("%v", raw.ResultCode) != "0" {
		return nil, fmt.Errorf("baidu: API error ResultCode=%v", raw.ResultCode)
	}

	out := &FundFlow{}

	// per-stock fund flow (fundFlowSpread)
	sp := raw.Result.Content.FundFlowSpread.Result
	if sp.Unit != "" {
		out.StockFlow = &StockFundFlow{
			OrderFlowGroups: OrderFlowGroups{
				Super:  mapGroup(sp.SuperGrp),
				Large:  mapGroup(sp.LargeGrp),
				Medium: mapGroup(sp.MediumGrp),
				Little: mapGroup(sp.LittleGrp),
			},
			TodayMain: MainFlow{
				MainIn:    sp.TodayMainFlow.MainIn,
				MainOut:   sp.TodayMainFlow.MainOut,
				MainNetIn: sp.TodayMainFlow.MainNetIn,
			},
			TurnoverInTotal:  sp.TurnoverInTotal,
			TurnoverOutTotal: sp.TurnoverOutTotal,
			TurnoverNetTotal: sp.TurnoverNetTotal,
			MainNetTotal:     sp.MainNetTotal,
			Unit:             sp.Unit,
			UpdateTime:       sp.UpdateTime,
			StockStatus:      sp.StockStatus,
		}
	}

	// industry-level fund flow (fundFlowBlock)
	for _, r := range raw.Result.Content.FundFlowBlock.Result {
		level := &FundFlowLevel{
			Belongs:    r.Belongs,
			UpdateTime: r.UpdateTime,
			Unit:       r.Unit,
			Industry:   IndustryInfo(r.Industry),
			TodayMain: MainFlow{
				MainIn:    r.TodayMainFlow.MainIn,
				MainOut:   r.TodayMainFlow.MainOut,
				MainNetIn: r.TodayMainFlow.MainNetIn,
			},
			OrderFlowGroups: OrderFlowGroups{
				Super:  mapGroup(r.SuperGrp),
				Large:  mapGroup(r.LargeGrp),
				Medium: mapGroup(r.MediumGrp),
				Little: mapGroup(r.LittleGrp),
			},
			Recently: r.Recently,
		}
		out.Levels = append(out.Levels, level)
	}
	return out, nil
}

type rawGroup struct {
	NetTurnover     any    `json:"netTurnover"`
	TurnoverIn      string `json:"turnoverIn"`
	TurnoverOut     string `json:"turnoverOut"`
	TurnoverInRate  string `json:"turnoverInRate"`
	TurnoverOutRate string `json:"turnoverOutRate"`
}

type rawMainFlow struct {
	MainIn    string `json:"mainIn"`
	MainOut   string `json:"mainOut"`
	MainNetIn string `json:"mainNetIn"`
}

type rawIndustry struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

func mapGroup(g rawGroup) OrderFlowGroup {
	return OrderFlowGroup{
		NetTurnover:     fmt.Sprintf("%v", g.NetTurnover),
		TurnoverIn:      g.TurnoverIn,
		TurnoverOut:     g.TurnoverOut,
		TurnoverInRate:  g.TurnoverInRate,
		TurnoverOutRate: g.TurnoverOutRate,
	}
}
