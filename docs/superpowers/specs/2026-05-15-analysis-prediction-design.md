# Analysis Prediction & History Design Spec

**Goal:** Complete the analysis package rewrite, add daily prediction tracking, and display predictions vs actual prices in CLI and Web.

**Architecture:** Backend calculates spread statistics per trading day using data up to the previous day, predicts the day's high/low/close based on that day's open price + "近2周" spread means, stores both predicted and actual values in `analysis_predictions` table. CLI and Web display the comparison.

---

## 1. Database: `analysis_predictions` table

One row per stock per trading day. Upsert on `(ts_code, trade_date)`.

| Field            | Type          | GORM tag                          | Description                              |
|------------------|---------------|-----------------------------------|------------------------------------------|
| id               | uint          | primaryKey                        |                                          |
| ts_code          | string(16)    | uniqueIndex:idx_code_date;not null| Stock code                               |
| trade_date       | string(8)     | uniqueIndex:idx_code_date;not null| YYYYMMDD                                 |
| sample_counts    | text          | type:json                         | `{"历史":2421,"近3月":90,"近1月":30,"近2周":15}` |
| window_means     | text          | type:json                         | Full `MeansData` JSON per window         |
| composite_means  | text          | type:json                         | `{"spread_oh":8.33,"spread_ol":5.54,...}` |
| open_price       | float64       |                                   | Actual open price for that day           |
| predict_high     | float64       |                                   | open + spread_oh (近2周 mean)             |
| predict_low      | float64       |                                   | open - spread_ol (近2周 mean)             |
| predict_close    | *float64      |                                   | Reserved, currently null (algorithm TBD) |
| actual_high      | *float64      |                                   | Actual high (null if market hasn't closed)|
| actual_low       | *float64      |                                   | Actual low                               |
| actual_close     | *float64      |                                   | Actual close                             |
| created_at       | time.Time     |                                   |                                          |
| updated_at       | time.Time     |                                   |                                          |

**Prediction formula** (using "近2周" window means):
- `predict_high = open_price + mean(spread_oh)`
- `predict_low = open_price - mean(spread_ol)`
- `predict_close` = null (algorithm TBD, reserved for future)

## 2. Calculation flow

### Recalc (CLI/Web triggered)

```
For each stock (or single stock if filtered):
  1. Load all daily_bars ORDER BY trade_date ASC
  2. For each bar at index i (starting from i >= 15 to ensure "近2周" has data):
     a. Historical slice = bars[0:i] (all bars before this day)
     b. This day's bar = bars[i]
     c. Call analysis.Build() with the historical slice
     d. Extract "近2周" window means
     e. predict_high = bars[i].Open + means.SpreadOH.Mean
     f. predict_low = bars[i].Open - means.SpreadOL.Mean
     g. predict_close = nil (TBD)
     h. actual_high/low/close = bars[i].High/Low/Close
     i. Upsert into analysis_predictions
```

### Incremental (after daily fetch)

Same as recalc but only for the latest trading day. Triggered automatically after `daily-fetch` job completes.

## 3. API endpoints

| Method | Path                           | Handler             | Description              |
|--------|--------------------------------|---------------------|--------------------------|
| POST   | /api/analysis/recalc           | RecalcAnalysis      | Trigger recalc; optional `?ts_code=xxx` |
| GET    | /api/analysis/predictions      | ListPredictions     | Query by `ts_code`; optional `from`/`to` date range |

### RecalcAnalysis request/response

```go
// Query params
type RecalcReq struct {
    TsCode string `form:"ts_code"` // optional, empty = all portfolio stocks
}

// Response
200 { "request_id": "...", "code": 200, "message": "ok", "data": { "updated": 42 } }
```

### ListPredictions request/response

```go
type ListPredictionsReq struct {
    TsCode string `form:"ts_code" binding:"required"`
    From   string `form:"from"`   // optional YYYYMMDD
    To     string `form:"to"`     // optional YYYYMMDD
    Limit  int    `form:"limit"`  // default 30
}
```

Returns array of `AnalysisPrediction` JSON.

## 4. Model struct

File: `pkg/models/analysis_prediction.go`

```go
type AnalysisPrediction struct {
    ID             uint            `gorm:"primaryKey" json:"id"`
    TsCode         string          `gorm:"uniqueIndex:idx_code_date;size:16;not null" json:"tsCode"`
    TradeDate      string          `gorm:"uniqueIndex:idx_code_date;size:8;not null" json:"tradeDate"`
    SampleCounts   StringJSONMap   `gorm:"type:json" json:"sampleCounts"`
    WindowMeans    StringJSONAny   `gorm:"type:json" json:"windowMeans"`
    CompositeMeans StringJSONMap   `gorm:"type:json" json:"compositeMeans"`
    OpenPrice      float64         `json:"openPrice"`
    PredictHigh    float64         `json:"predictHigh"`
    PredictLow     float64         `json:"predictLow"`
    PredictClose   *float64        `json:"predictClose"`
    ActualHigh     *float64        `json:"actualHigh"`
    ActualLow      *float64        `json:"actualLow"`
    ActualClose    *float64        `json:"actualClose"`
    CreatedAt      time.Time       `json:"createdAt"`
    UpdatedAt      time.Time       `json:"updatedAt"`
}
```

`StringJSONMap` = `map[string]float64` with custom JSON scan/value.
`StringJSONAny` = `json.RawMessage` for flexible nested structure.

## 5. CLI commands

### `stockctl analysis recalc [--code 603778]`

- Calls `POST /api/analysis/recalc?ts_code=xxx`
- Prints: `Recalculated 42 predictions for 603778`

### `stockctl analysis predictions <code>`

- Calls `GET /api/analysis/predictions?ts_code=xxx`
- Prints ASCII table matching Python style:

```
603778 (中国交建) 预测记录

日期       | 预测高   | 实际高   | 偏差     | 预测低   | 实际低   | 偏差
20250513  | 356.22  | 355.80  | -0.42   | 353.76  | 354.10  | +0.34
20250512  | 354.50  | 354.83  | +0.33   | 352.80  | 353.20  | +0.40
```

## 6. Web frontend

### Predictions tab in StockDetailView

Add a third tab "预测记录" alongside "基础与价差" and "详细统计".

Content: el-table showing the same data as CLI, with `g-up`/`g-down` coloring on deviation columns.

### Recalc trigger

Add a "重新计算" button in the Predictions tab header. Calls `POST /api/analysis/recalc?ts_code=xxx`.

## 7. Builder completion

The `Build()` function in `builder.go` must be completed to produce:

1. **ModelTable** - Spread model (4 windows × 6 spreads + composite row)
2. **ReferenceTable** - Price prediction reference (high/low/close × 4 windows)
3. **AnalysisTable** (new field on AnalysisResult) - OH+OL statistics + recommendation ranges

New function needed:
- `Composite(windows []*WindowData) map[string]float64` - Average across all windows for each spread key
- `buildAnalysisTable(windows)` - Build the OH+OL stats + recommend ranges table
- `recommendRange(sorted []float64, threshold float64) (low, high float64, cumPct float64)` - Sliding window algorithm from Python

## 8. Auto-incremental after daily fetch

In the scheduler's `daily-fetch` handler, after bars are synced, call the recalc logic for each portfolio stock with only the latest trading day. This ensures predictions are generated automatically.

---

## Scope

**In scope:**
- Complete `builder.go` (Build, Composite, ModelTable, ReferenceTable, AnalysisTable)
- New `analysis_predictions` table + model
- New `pkg/stockd/services/prediction` service (recalc + query)
- New API endpoints (recalc + list predictions)
- New CLI commands (recalc + predictions)
- CLI render for prediction table
- Web: predictions tab + recalc button
- Auto-incremental after daily-fetch

**Out of scope:**
- Distribution table display (data is stored but not rendered)
- Full 6-spread analysis view (only default OH+OL view)
- Daily data table (not needed per earlier decision)
