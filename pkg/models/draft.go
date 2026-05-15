package models

type UpsertDraftReq struct {
	TsCode    string   `json:"tsCode"`
	TradeDate string   `json:"tradeDate"`
	Open      *float64 `json:"open,omitempty"`
	High      *float64 `json:"high,omitempty"`
	Low       *float64 `json:"low,omitempty"`
	Close     *float64 `json:"close,omitempty"`
}
