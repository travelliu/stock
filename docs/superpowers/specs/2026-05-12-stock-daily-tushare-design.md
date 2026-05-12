# Stock Daily Data System (Tushare)

## Overview

Replace the existing akshare-based stock analysis system with a tushare-based daily data system. Fetch daily OHLCV data for 8 specified stocks (last 6 months), store in SQLite, and provide a unified CLI for data retrieval and price-spread analysis.

## Requirements

- Data source: tushare `pro.daily()` API
- Storage: SQLite (single-file database)
- Stocks: 600537, 603778, 000890, 002709, 600186, 002842, 300342, 000593
- Date range: last 180 days (approximately 6 months)
- CLI: unified entry point with subcommands (argparse)
- Analysis: pre-compute 6 price-spread types during fetch, store in DB
- Display: show command with optional `--analyze` flag for statistics

## Architecture

### File Structure

```
stock/
├── stock.py            # CLI entry point (argparse subcommands)
├── db.py               # SQLite database operations (schema, CRUD)
├── fetcher.py          # tushare data fetching logic
├── analysis.py         # Price-spread statistical analysis
├── config.py           # Global configuration
├── requirements.txt    # Dependencies: tushare, pandas, tabulate
├── data/
│   └── stock_daily.db  # SQLite database file
```

### Deleted Files

- `init_data.py`
- `morning_scan.py`
- `evening_close.py`
- `fetch_minute_auction.py`
- `analyzer.py`

### Module Responsibilities

| Module | Responsibility |
|--------|---------------|
| `stock.py` | CLI entry point, parse arguments, dispatch to fetch/show subcommands |
| `db.py` | SQLite operations: initialize schema, insert daily data with spreads, query data |
| `fetcher.py` | Call tushare API, fetch daily OHLCV, return DataFrame |
| `analysis.py` | Calculate mean and median for each spread type over a date range |
| `config.py` | tushare token, DB path, default stock list, fetch parameters |

### Data Flow

```
tushare API → fetcher.py → compute spreads → db.py (write SQLite)
                                                    ↓
CLI show ← db.py (read) ← SQLite
CLI show --analyze ← analysis.py ← db.py (read) ← SQLite
```

## Database Schema

```sql
CREATE TABLE IF NOT EXISTS daily (
    ts_code     TEXT NOT NULL,       -- tushare format: '600537.SH', '000890.SZ'
    trade_date  TEXT NOT NULL,       -- '2025-01-02'
    open        REAL NOT NULL,       -- opening price
    high        REAL NOT NULL,       -- highest price
    low         REAL NOT NULL,       -- lowest price
    close       REAL NOT NULL,       -- closing price
    vol         REAL NOT NULL,       -- volume (lots)
    amount      REAL NOT NULL,       -- turnover (thousand yuan)
    spread_oh   REAL,               -- high - open
    spread_ol   REAL,               -- open - low
    spread_hl   REAL,               -- high - low (amplitude)
    spread_oc   REAL,               -- open - close
    spread_hc   REAL,               -- high - close
    spread_lc   REAL,               -- low - close
    PRIMARY KEY (ts_code, trade_date)
);
```

- `ts_code` uses tushare format: `6xx` → `.SH`, `0xx/3xx` → `.SZ`
- Primary key `(ts_code, trade_date)` prevents duplicates
- `INSERT OR REPLACE` for idempotent writes
- Spread columns are pre-computed during fetch

## CLI Design

### Commands

```
python stock.py fetch [--stocks CODE,...] [--days N]
python stock.py show --stock CODE [--from DATE] [--to DATE] [--analyze]
```

### `fetch` Subcommand

- `--stocks`: comma-separated stock codes (default: all 8 configured stocks)
- `--days`: lookback days from today (default: 180)
- Behavior:
  1. Convert user codes to tushare format (e.g., `600537` → `600537.SH`)
  2. Call `pro.daily()` for each stock with date range
  3. Compute 6 spread values per row
  4. Insert into SQLite with `INSERT OR REPLACE`

### `show` Subcommand

- `--stock`: single stock code (required)
- `--from`: start date, default 30 days ago
- `--to`: end date, default today
- `--analyze`: flag, when present show spread statistics instead of raw data

Without `--analyze`: display OHLCV + spread columns in table format.

With `--analyze`: display mean and median for each spread type.

### Analyze Output Format

```
=== 600517 价差分析 (2025-01-01 ~ 2025-05-12) ===
样本数: 98

价差类型         平均值     中位数
─────────────────────────────────
最高-开盘        0.35      0.28
开盘-最低        0.22      0.18
最高-最低        0.57      0.48
开盘-收盘        0.08      0.05
最高-收盘        0.27      0.21
最低-收盘       -0.31     -0.25
```

## Configuration

```python
# config.py
TUSHARE_TOKEN = "YOUR_TUSHARE_TOKEN_HERE"
DB_PATH = DATA_DIR / "stock_daily.db"
DEFAULT_STOCKS = ["600537", "603778", "000890", "002709", "600186", "002842", "300342", "000593"]
DEFAULT_FETCH_DAYS = 180
```

## Dependencies

```
tushare>=1.4.0
pandas>=2.0.0
tabulate>=0.9.0
```

## Stock Code Mapping

### Market Determination Rules

| Code Prefix | Market | Board |
|-------------|--------|-------|
| 600 / 601 / 603 / 605 | SH (上交所) | 主板A股 |
| 688 | SH (上交所) | 科创板 |
| 900 | SH (上交所) | B股 |
| 000 | SZ (深交所) | 主板A股 |
| 001 | SZ (深交所) | 主板（近年新增） |
| 002 | SZ (深交所) | 中小板 |
| 300 | SZ (深交所) | 创业板 |
| 200 | SZ (深交所) | B股 |
| 510 / 511 / 512 / 513 / 515 | SH (上交所) | ETF |
| 159 | SZ (深交所) | ETF |

IPO subscription codes (730, 732, 00x etc.) are not real trading codes and will raise `ValueError`.

### Default Stock Examples

| User Input | Tushare Format | Market | Board |
|-----------|---------------|--------|-------|
| 600537 | 600537.SH | Shanghai | 主板A股 |
| 603778 | 603778.SH | Shanghai | 主板A股 |
| 000890 | 000890.SZ | Shenzhen | 主板A股 |
| 002709 | 002709.SZ | Shenzhen | 中小板 |
| 600186 | 600186.SH | Shanghai | 主板A股 |
| 002842 | 002842.SZ | Shenzhen | 中小板 |
| 300342 | 300342.SZ | Shenzhen | 创业板 |
| 000593 | 000593.SZ | Shenzhen | 主板A股 |

### Algorithm

```
if code already contains '.' → pass through
prefix = first 3 digits of code

SH prefixes: 600, 601, 603, 605, 688, 900, 510, 511, 512, 513, 515
SZ prefixes: 000, 001, 002, 300, 200, 159

if prefix in SH → code + ".SH"
elif prefix in SZ → code + ".SZ"
else → raise ValueError
```
