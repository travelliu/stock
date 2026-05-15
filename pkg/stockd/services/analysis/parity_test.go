package analysis_test

type pythonRow struct {
	TsCode    string  `json:"ts_code"`
	TradeDate string  `json:"trade_date"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Vol       float64 `json:"vol"`
	Amount    float64 `json:"amount"`
	SpreadOH  float64 `json:"spread_oh"`
	SpreadOL  float64 `json:"spread_ol"`
	SpreadHL  float64 `json:"spread_hl"`
	SpreadOC  float64 `json:"spread_oc"`
	SpreadHC  float64 `json:"spread_hc"`
	SpreadLC  float64 `json:"spread_lc"`
}

type pythonFixture struct {
	Code               string                         `json:"code"`
	Rows               []pythonRow                    `json:"rows"`
	OpenPrice          float64                        `json:"open_price"`
	ActualHigh         *float64                       `json:"actual_high"`
	ActualLow          *float64                       `json:"actual_low"`
	ActualClose        *float64                       `json:"actual_close"`
	WindowMeans        map[string]map[string]*float64 `json:"window_means"`
	CompositeMeans     map[string]float64             `json:"composite_means"`
	ModelTableText     string                         `json:"model_table_text"`
	ReferenceTableText string                         `json:"reference_table_text"`
}

// func TestParity_AgainstPythonFixtures(t *testing.T) {
// 	matches, err := filepath.Glob("testdata/*.json")
// 	require.NoError(t, err)
// 	require.NotEmpty(t, matches, "no parity fixtures generated yet (see tools/dump_python_fixture.py)")
//
// 	for _, path := range matches {
// 		t.Run(filepath.Base(path), func(t *testing.T) {
// 			raw, err := os.ReadFile(path)
// 			require.NoError(t, err)
// 			var fx pythonFixture
// 			require.NoError(t, json.Unmarshal(raw, &fx))
//
// 			bars := make([]models.DailyBar, 0, len(fx.Rows))
// 			for _, r := range fx.Rows {
// 				bars = append(bars, models.DailyBar{
// 					TsCode: r.TsCode, TradeDate: r.TradeDate,
// 					Open: r.Open, High: r.High, Low: r.Low, Close: r.Close,
// 					Vol: r.Vol, Amount: r.Amount,
// 					Spreads: models.Spreads{
// 						OH: r.SpreadOH, OL: r.SpreadOL, HL: r.SpreadHL,
// 						OC: r.SpreadOC, HC: r.SpreadHC, LC: r.SpreadLC,
// 					},
// 				})
// 			}
// 			in := models.Input{
// 				TsCode:      fx.Code,
// 				Rows:        bars,
// 				OpenPrice:   ptrFloatIf(fx.OpenPrice),
// 				ActualHigh:  fx.ActualHigh,
// 				ActualLow:   fx.ActualLow,
// 				ActualClose: fx.ActualClose,
// 			}
// 			res := Build(in)
//
// 			// Composite means must match to 4 decimal places.
// 			for _, key := range ModelSpreadKeys {
// 				assert.InDelta(t, fx.CompositeMeans[key], res.CompositeMeans[key], 1e-4, "composite[%s]", key)
// 			}
//
// 			// WindowData means: nil-equal and value-equal.
// 			for w, byKey := range fx.WindowMeans {
// 				for k, want := range byKey {
// 					got := res.WindowMeans[w][k]
// 					if want == nil {
// 						assert.Nil(t, got, "%s/%s should be nil", w, k)
// 						continue
// 					}
// 					require.NotNil(t, got, "%s/%s should not be nil", w, k)
// 					assert.InDelta(t, *want, *got, 1e-4)
// 				}
// 			}
//
// 			// Rendered model table must equal byte-for-byte.
// 			goModel := analysis.FormatTable(res.ModelTable.Headers, res.ModelTable.Rows)
// 			assert.Equal(t, fx.ModelTableText, goModel, "model_table rendering differs for %s", fx.Code)
//
// 			if fx.OpenPrice != 0 {
// 				goRef := analysis.FormatTable(res.ReferenceTable.Headers, res.ReferenceTable.Rows)
// 				assert.Equal(t, fx.ReferenceTableText, goRef, "reference_table rendering differs for %s", fx.Code)
// 			}
// 		})
// 	}
// }
//
// func ptrFloatIf(v float64) *float64 {
// 	if v == 0 {
// 		return nil
// 	}
// 	return &v
// }
