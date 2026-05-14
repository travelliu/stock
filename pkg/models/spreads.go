package models

type Spreads struct {
	OH float64 `json:"oh,omitempty"`
	OL float64 `json:"ol,omitempty"`
	HL float64 `json:"hl,omitempty"`
	OC float64 `json:"oc,omitempty"`
	HC float64 `json:"hc,omitempty"`
	LC float64 `json:"lc,omitempty"`
}
