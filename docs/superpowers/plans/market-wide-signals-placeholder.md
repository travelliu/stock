# 全市场信号层（占位，待议）

> **状态：** 暂不实现，等个股阶段完成后再评估。

**Goal:** 接通全市场信号端点：THS 热股归因、北向资金、行业排名、龙虎榜（全市场）、财联社快讯。

**注意：** 这些端点与个股无关，面向全市场，技术复杂度（5 分钟 TTL 缓存、并发刷新、数据量）和业务价值需单独评估后再规划。

---

## 待规划端点

| Method | Path | Source | Description |
|--------|------|--------|-------------|
| GET | `/api/signals/hot` | `zx.10jqka.com.cn` | 每日强势股 + 题材原因标签 |
| GET | `/api/signals/northbound` | `data.hexin.cn` | 北向资金分钟流向（沪股通/深股通） |
| GET | `/api/signals/industry` | 同花顺行业 API | ~90 个行业按涨跌幅/净流入/龙头排序 |
| GET | `/api/signals/dragon-tiger` | 东方财富 datacenter | 全市场龙虎榜（按日期查询 `?date=YYYY-MM-DD`） |
| GET | `/api/signals/cls-news` | 财联社 API | 实时市场快讯 |

## 技术要点

- **5 分钟 TTL 缓存：** 复用 `realtimeCache` 模式，加 `signalCache map[string]signalEntry` + `signalMu sync.RWMutex`，entry 含 `data any` + `expiredAt time.Time`。
- **新包：** `pkg/ths/`（热股 + 行业，EPS 在 Phase 4 已建）、`pkg/hexin/`（北向）、扩展 `pkg/eastmoney/`（全市场龙虎）。
- **财联社：** URL 需逆向工程（参考 akshare 源码）。
- **handler 文件：** `pkg/stockd/http/signals.go`（新建）。

## 依赖前提

- 个股 Phase 3 完成（`pkg/eastmoney/` 已建）
- 个股 Phase 4 完成（`pkg/ths/` 已建，只需新增热股/行业方法）
