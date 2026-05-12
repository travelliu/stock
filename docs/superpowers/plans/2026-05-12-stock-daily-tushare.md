# Stock Daily Data System (Tushare) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the akshare-based system with a tushare daily data system — fetch OHLCV + spreads, store in SQLite, provide CLI for retrieval and analysis.

**Architecture:** Single-package Python CLI with 5 modules. tushare `pro.daily()` feeds data into SQLite. Spreads are pre-computed at fetch time. CLI uses argparse subcommands (`fetch`, `show`). Test stock: **603778**.

**Tech Stack:** Python 3.10, tushare, pandas, tabulate, SQLite3 (stdlib), pytest

---

## File Structure

```
stock/
├── config.py           # tushare token, DB path, stock list, helpers
├── db.py               # SQLite schema, insert, query
├── fetcher.py          # tushare API fetch + spread computation
├── analysis.py         # mean/median statistics over date range
├── stock.py            # CLI entry point (argparse)
├── requirements.txt    # updated: tushare replaces akshare
├── data/               # auto-created at runtime
│   └── stock_daily.db
├── tests/
│   ├── test_config.py
│   ├── test_db.py
│   ├── test_fetcher.py
│   ├── test_analysis.py
│   └── test_cli.py
```

**Deleted files** (already removed from working tree):
`init_data.py`, `morning_scan.py`, `evening_close.py`, `analyzer.py`, `config.py` (old version)

---

## Task 1: config.py — Configuration & Stock Code Helpers

**Files:**
- Create: `config.py`
- Create: `tests/test_config.py`

- [ ] **Step 1: Write the failing test**

```python
# tests/test_config.py
import pytest
from config import to_tushare_code, TUSHARE_TOKEN, DB_PATH, DEFAULT_STOCKS


class TestToTushareCode:
    # Shanghai main board
    def test_shanghai_600(self):
        assert to_tushare_code("600537") == "600537.SH"

    def test_shanghai_601(self):
        assert to_tushare_code("601398") == "601398.SH"

    def test_shanghai_603(self):
        assert to_tushare_code("603778") == "603778.SH"

    def test_shanghai_605(self):
        assert to_tushare_code("605123") == "605123.SH"

    # Shanghai STAR (科创板)
    def test_shanghai_688(self):
        assert to_tushare_code("688001") == "688001.SH"

    # Shanghai B-share
    def test_shanghai_900(self):
        assert to_tushare_code("900901") == "900901.SH"

    # Shenzhen main board
    def test_shenzhen_000(self):
        assert to_tushare_code("000890") == "000890.SZ"

    def test_shenzhen_001(self):
        assert to_tushare_code("001234") == "001234.SZ"

    # Shenzhen SME (中小板)
    def test_shenzhen_002(self):
        assert to_tushare_code("002709") == "002709.SZ"

    # Shenzhen ChiNext (创业板)
    def test_shenzhen_300(self):
        assert to_tushare_code("300342") == "300342.SZ"

    # Shenzhen B-share
    def test_shenzhen_200(self):
        assert to_tushare_code("200002") == "200002.SZ"

    # Shanghai ETF
    def test_shanghai_etf_510(self):
        assert to_tushare_code("510050") == "510050.SH"

    def test_shanghai_etf_512(self):
        assert to_tushare_code("512100") == "512100.SH"

    def test_shanghai_etf_515(self):
        assert to_tushare_code("515030") == "515030.SH"

    # Shenzhen ETF
    def test_shenzhen_etf_159(self):
        assert to_tushare_code("159915") == "159915.SZ"

    # Pass-through for already-suffixed codes
    def test_already_suffixed_sh(self):
        assert to_tushare_code("600537.SH") == "600537.SH"

    def test_already_suffixed_sz(self):
        assert to_tushare_code("000890.SZ") == "000890.SZ"

    # Invalid / IPO subscription codes
    def test_ipo_730_raises(self):
        with pytest.raises(ValueError, match="Cannot determine market"):
            to_tushare_code("730123")

    def test_unknown_prefix_raises(self):
        with pytest.raises(ValueError, match="Cannot determine market"):
            to_tushare_code("500123")

    def test_4digit_short_code_raises(self):
        with pytest.raises(ValueError, match="must be a 6-digit code"):
            to_tushare_code("60")


class TestConfig:
    def test_token_not_empty(self):
        assert TUSHARE_TOKEN

    def test_db_path_is_absolute(self):
        assert DB_PATH.is_absolute()

    def test_default_stocks_has_8(self):
        assert len(DEFAULT_STOCKS) == 8

    def test_default_stocks_are_plain_codes(self):
        for code in DEFAULT_STOCKS:
            assert "." not in code
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /root/code/github/travelliu/stock && python -m pytest tests/test_config.py -v`
Expected: FAIL — `ModuleNotFoundError: No module named 'config'`

- [ ] **Step 3: Write minimal implementation**

```python
# config.py
from pathlib import Path

# Directories
PROJECT_DIR = Path(__file__).resolve().parent
DATA_DIR = PROJECT_DIR / "data"
DATA_DIR.mkdir(exist_ok=True)

# tushare
TUSHARE_TOKEN = "YOUR_TUSHARE_TOKEN_HERE"

# Database
DB_PATH = DATA_DIR / "stock_daily.db"

# Default stocks
DEFAULT_STOCKS = [
    "600537", "603778", "000890", "002709",
    "600186", "002842", "300342", "000593",
]

# Fetch defaults
DEFAULT_FETCH_DAYS = 180

# Market determination by 3-digit prefix
_SH_PREFIXES = {"600", "601", "603", "605", "688", "900",
                "510", "511", "512", "513", "515"}
_SZ_PREFIXES = {"000", "001", "002", "300", "200", "159"}


def to_tushare_code(code: str) -> str:
    """Convert plain stock code to tushare format.

    Rules:
        SH (上交所): 600/601/603/605 (主板), 688 (科创板),
                     900 (B股), 510/511/512/513/515 (ETF)
        SZ (深交所): 000/001 (主板), 002 (中小板), 300 (创业板),
                     200 (B股), 159 (ETF)
        IPO subscription codes (730, 732, etc.) raise ValueError.

    Examples:
        '600537' -> '600537.SH'
        '000890' -> '000890.SZ'
        '688001' -> '688001.SH'
        '159915' -> '159915.SZ'

    Already-suffixed codes pass through unchanged.
    """
    if "." in code:
        return code
    if len(code) < 6:
        raise ValueError(
            f"Invalid stock code '{code}': must be a 6-digit code"
        )
    prefix = code[:3]
    if prefix in _SH_PREFIXES:
        return f"{code}.SH"
    if prefix in _SZ_PREFIXES:
        return f"{code}.SZ"
    raise ValueError(
        f"Cannot determine market for stock code '{code}' "
        f"(prefix '{prefix}' is not a known SH/SZ prefix)"
    )
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /root/code/github/travelliu/stock && python -m pytest tests/test_config.py -v`
Expected: all 22 tests PASS

- [ ] **Step 5: Commit**

```bash
git add config.py tests/test_config.py
git commit -m "feat: add config module with stock code mapping"
```

---

## Task 2: db.py — SQLite Schema & Data Operations

**Files:**
- Create: `db.py`
- Create: `tests/test_db.py`

- [ ] **Step 1: Write the failing test**

```python
# tests/test_db.py
import sqlite3
import tempfile
from pathlib import Path

import pandas as pd
import pytest

from db import DailyDB


@pytest.fixture
def db(tmp_path):
    db_path = tmp_path / "test.db"
    database = DailyDB(db_path)
    database.init()
    return database


@pytest.fixture
def sample_df():
    return pd.DataFrame([
        {
            "ts_code": "603778.SH",
            "trade_date": "2025-05-12",
            "open": 10.50,
            "high": 11.20,
            "low": 10.30,
            "close": 11.00,
            "vol": 50000.0,
            "amount": 550000.0,
            "spread_oh": 0.70,
            "spread_ol": 0.20,
            "spread_hl": 0.90,
            "spread_oc": -0.50,
            "spread_hc": 0.20,
            "spread_lc": -0.70,
        },
    ])


class TestInit:
    def test_creates_table(self, db):
        conn = sqlite3.connect(db.db_path)
        cursor = conn.execute(
            "SELECT name FROM sqlite_master WHERE type='table' AND name='daily'"
        )
        assert cursor.fetchone() is not None
        conn.close()

    def test_idempotent(self, db):
        db.init()
        conn = sqlite3.connect(db.db_path)
        cursor = conn.execute(
            "SELECT name FROM sqlite_master WHERE type='table' AND name='daily'"
        )
        assert cursor.fetchone() is not None
        conn.close()


class TestInsert:
    def test_insert_one_row(self, db, sample_df):
        db.insert_daily(sample_df)
        rows = db.query_daily("603778", "2025-05-01", "2025-05-31")
        assert len(rows) == 1
        assert rows[0]["ts_code"] == "603778.SH"
        assert rows[0]["close"] == 11.00

    def test_insert_replace_upsert(self, db, sample_df):
        db.insert_daily(sample_df)
        modified = sample_df.copy()
        modified["close"] = 12.00
        db.insert_daily(modified)
        rows = db.query_daily("603778", "2025-05-01", "2025-05-31")
        assert len(rows) == 1
        assert rows[0]["close"] == 12.00


class TestQuery:
    def test_query_by_date_range(self, db):
        df = pd.DataFrame([
            {"ts_code": "603778.SH", "trade_date": "2025-05-10",
             "open": 10.0, "high": 11.0, "low": 9.5, "close": 10.5,
             "vol": 1000.0, "amount": 10000.0,
             "spread_oh": 1.0, "spread_ol": 0.5, "spread_hl": 1.5,
             "spread_oc": -0.5, "spread_hc": 0.5, "spread_lc": -1.0},
            {"ts_code": "603778.SH", "trade_date": "2025-05-12",
             "open": 11.0, "high": 12.0, "low": 10.5, "close": 11.5,
             "vol": 2000.0, "amount": 20000.0,
             "spread_oh": 1.0, "spread_ol": 0.5, "spread_hl": 1.5,
             "spread_oc": -0.5, "spread_hc": 0.5, "spread_lc": -1.0},
            {"ts_code": "603778.SH", "trade_date": "2025-05-15",
             "open": 12.0, "high": 13.0, "low": 11.5, "close": 12.5,
             "vol": 3000.0, "amount": 30000.0,
             "spread_oh": 1.0, "spread_ol": 0.5, "spread_hl": 1.5,
             "spread_oc": -0.5, "spread_hc": 0.5, "spread_lc": -1.0},
        ])
        db.insert_daily(df)
        rows = db.query_daily("603778", "2025-05-11", "2025-05-14")
        dates = [r["trade_date"] for r in rows]
        assert "2025-05-10" not in dates
        assert "2025-05-12" in dates
        assert "2025-05-15" not in dates

    def test_query_returns_spread_columns(self, db, sample_df):
        db.insert_daily(sample_df)
        rows = db.query_daily("603778", "2025-05-01", "2025-05-31")
        row = rows[0]
        assert row["spread_oh"] == 0.70
        assert row["spread_hl"] == 0.90
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /root/code/github/travelliu/stock && python -m pytest tests/test_db.py -v`
Expected: FAIL — `ModuleNotFoundError: No module named 'db'`

- [ ] **Step 3: Write minimal implementation**

```python
# db.py
import sqlite3
from pathlib import Path
from typing import Any

import pandas as pd

SCHEMA = """
CREATE TABLE IF NOT EXISTS daily (
    ts_code     TEXT NOT NULL,
    trade_date  TEXT NOT NULL,
    open        REAL NOT NULL,
    high        REAL NOT NULL,
    low         REAL NOT NULL,
    close       REAL NOT NULL,
    vol         REAL NOT NULL,
    amount      REAL NOT NULL,
    spread_oh   REAL,
    spread_ol   REAL,
    spread_hl   REAL,
    spread_oc   REAL,
    spread_hc   REAL,
    spread_lc   REAL,
    PRIMARY KEY (ts_code, trade_date)
);
"""

COLUMNS = [
    "ts_code", "trade_date", "open", "high", "low", "close",
    "vol", "amount",
    "spread_oh", "spread_ol", "spread_hl",
    "spread_oc", "spread_hc", "spread_lc",
]


class DailyDB:
    def __init__(self, db_path: Path) -> None:
        self.db_path = db_path

    def init(self) -> None:
        conn = sqlite3.connect(self.db_path)
        conn.execute(SCHEMA)
        conn.commit()
        conn.close()

    def insert_daily(self, df: pd.DataFrame) -> int:
        conn = sqlite3.connect(self.db_path)
        placeholders = ", ".join(["?"] * len(COLUMNS))
        cols = ", ".join(COLUMNS)
        sql = f"INSERT OR REPLACE INTO daily ({cols}) VALUES ({placeholders})"
        rows = df[COLUMNS].values.tolist()
        conn.executemany(sql, rows)
        count = conn.total_changes
        conn.commit()
        conn.close()
        return count

    def query_daily(
        self,
        stock_code: str,
        start_date: str,
        end_date: str,
    ) -> list[dict[str, Any]]:
        from config import to_tushare_code

        ts_code = to_tushare_code(stock_code)
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        cursor = conn.execute(
            "SELECT * FROM daily "
            "WHERE ts_code = ? AND trade_date >= ? AND trade_date <= ? "
            "ORDER BY trade_date",
            (ts_code, start_date, end_date),
        )
        rows = [dict(row) for row in cursor.fetchall()]
        conn.close()
        return rows
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /root/code/github/travelliu/stock && python -m pytest tests/test_db.py -v`
Expected: all 6 tests PASS

- [ ] **Step 5: Commit**

```bash
git add db.py tests/test_db.py
git commit -m "feat: add db module with SQLite schema and CRUD"
```

---

## Task 3: fetcher.py — Tushare Data Fetching & Spread Computation

**Files:**
- Create: `fetcher.py`
- Create: `tests/test_fetcher.py`

- [ ] **Step 1: Write the failing test**

```python
# tests/test_fetcher.py
import pytest
import pandas as pd
from unittest.mock import patch, MagicMock

from fetcher import compute_spreads, fetch_daily


class TestComputeSpreads:
    def test_basic_spread_computation(self):
        df = pd.DataFrame([
            {"open": 10.0, "high": 11.0, "low": 9.5, "close": 10.5},
        ])
        result = compute_spreads(df)
        row = result.iloc[0]
        assert row["spread_oh"] == pytest.approx(1.0)    # high - open
        assert row["spread_ol"] == pytest.approx(0.5)    # open - low
        assert row["spread_hl"] == pytest.approx(1.5)    # high - low
        assert row["spread_oc"] == pytest.approx(-0.5)   # open - close
        assert row["spread_hc"] == pytest.approx(0.5)    # high - close
        assert row["spread_lc"] == pytest.approx(-1.0)   # low - close

    def test_preserves_original_columns(self):
        df = pd.DataFrame([
            {"ts_code": "603778.SH", "trade_date": "2025-05-12",
             "open": 10.0, "high": 11.0, "low": 9.5, "close": 10.5,
             "vol": 1000.0, "amount": 10000.0},
        ])
        result = compute_spreads(df)
        assert "ts_code" in result.columns
        assert "trade_date" in result.columns
        assert result.iloc[0]["ts_code"] == "603778.SH"

    def test_multiple_rows(self):
        df = pd.DataFrame([
            {"open": 10.0, "high": 11.0, "low": 9.5, "close": 10.5},
            {"open": 20.0, "high": 21.0, "low": 19.0, "close": 20.5},
        ])
        result = compute_spreads(df)
        assert len(result) == 2
        assert result.iloc[1]["spread_oh"] == pytest.approx(1.0)


class TestFetchDaily:
    @patch("fetcher.tushare")
    def test_fetch_calls_api_with_correct_params(self, mock_ts):
        mock_pro = MagicMock()
        mock_ts.pro_api.return_value = mock_pro
        mock_pro.daily.return_value = pd.DataFrame([
            {"ts_code": "603778.SH", "trade_date": "20250512",
             "open": 10.0, "high": 11.0, "low": 9.5, "close": 10.5,
             "vol": 1000.0, "amount": 10000.0},
        ])
        result = fetch_daily("603778", days=30)
        mock_pro.daily.assert_called_once()
        call_kwargs = mock_pro.daily.call_args[1]
        assert call_kwargs["ts_code"] == "603778.SH"

    @patch("fetcher.tushare")
    def test_fetch_converts_date_format(self, mock_ts):
        mock_pro = MagicMock()
        mock_ts.pro_api.return_value = mock_pro
        mock_pro.daily.return_value = pd.DataFrame([
            {"ts_code": "603778.SH", "trade_date": "20250512",
             "open": 10.0, "high": 11.0, "low": 9.5, "close": 10.5,
             "vol": 1000.0, "amount": 10000.0},
        ])
        result = fetch_daily("603778", days=30)
        assert result.iloc[0]["trade_date"] == "2025-05-12"
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /root/code/github/travelliu/stock && python -m pytest tests/test_fetcher.py -v`
Expected: FAIL — `ModuleNotFoundError: No module named 'fetcher'`

- [ ] **Step 3: Write minimal implementation**

```python
# fetcher.py
from datetime import datetime, timedelta

import pandas as pd
import tushare

from config import TUSHARE_TOKEN, to_tushare_code


def compute_spreads(df: pd.DataFrame) -> pd.DataFrame:
    """Compute 6 price-spread columns and append to DataFrame.

    spread_oh = high - open
    spread_ol = open - low
    spread_hl = high - low
    spread_oc = open - close
    spread_hc = high - close
    spread_lc = low - close
    """
    result = df.copy()
    result["spread_oh"] = result["high"] - result["open"]
    result["spread_ol"] = result["open"] - result["low"]
    result["spread_hl"] = result["high"] - result["low"]
    result["spread_oc"] = result["open"] - result["close"]
    result["spread_hc"] = result["high"] - result["close"]
    result["spread_lc"] = result["low"] - result["close"]
    return result


def fetch_daily(stock_code: str, days: int = 180) -> pd.DataFrame:
    """Fetch daily OHLCV from tushare and compute spreads.

    Args:
        stock_code: plain code like '603778'
        days: lookback days from today

    Returns:
        DataFrame with OHLCV + spread columns, trade_date in 'YYYY-MM-DD' format.
    """
    ts_code = to_tushare_code(stock_code)
    end_date = datetime.now().strftime("%Y%m%d")
    start_date = (datetime.now() - timedelta(days=days)).strftime("%Y%m%d")

    pro = tushare.pro_api(TUSHARE_TOKEN)
    df = pro.daily(ts_code=ts_code, start_date=start_date, end_date=end_date)

    if df is None or df.empty:
        return pd.DataFrame()

    # Convert tushare date format 'YYYYMMDD' -> 'YYYY-MM-DD'
    df["trade_date"] = pd.to_datetime(df["trade_date"], format="%Y%m%d").dt.strftime(
        "%Y-%m-%d"
    )

    # Sort by date ascending
    df = df.sort_values("trade_date").reset_index(drop=True)

    return compute_spreads(df)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /root/code/github/travelliu/stock && python -m pytest tests/test_fetcher.py -v`
Expected: all 4 tests PASS

- [ ] **Step 5: Commit**

```bash
git add fetcher.py tests/test_fetcher.py
git commit -m "feat: add fetcher module with tushare API and spread computation"
```

---

## Task 4: analysis.py — Price Spread Statistics

**Files:**
- Create: `analysis.py`
- Create: `tests/test_analysis.py`

- [ ] **Step 1: Write the failing test**

```python
# tests/test_analysis.py
import pytest
from analysis import compute_statistics, SPREAD_LABELS


class TestComputeStatistics:
    def test_basic_statistics(self):
        rows = [
            {"spread_oh": 1.0, "spread_ol": 0.5, "spread_hl": 1.5,
             "spread_oc": -0.5, "spread_hc": 0.5, "spread_lc": -1.0},
            {"spread_oh": 2.0, "spread_ol": 1.0, "spread_hl": 3.0,
             "spread_oc": -1.0, "spread_hc": 1.0, "spread_lc": -2.0},
            {"spread_oh": 0.5, "spread_ol": 0.25, "spread_hl": 0.75,
             "spread_oc": -0.25, "spread_hc": 0.25, "spread_lc": -0.5},
        ]
        stats = compute_statistics(rows)
        assert stats["count"] == 3
        assert stats["spreads"]["spread_oh"]["mean"] == pytest.approx(1.1667, rel=1e-3)
        assert stats["spreads"]["spread_oh"]["median"] == pytest.approx(1.0)

    def test_empty_rows(self):
        stats = compute_statistics([])
        assert stats["count"] == 0
        assert stats["spreads"] == {}

    def test_single_row(self):
        rows = [
            {"spread_oh": 1.0, "spread_ol": 0.5, "spread_hl": 1.5,
             "spread_oc": -0.5, "spread_hc": 0.5, "spread_lc": -1.0},
        ]
        stats = compute_statistics(rows)
        assert stats["count"] == 1
        assert stats["spreads"]["spread_oh"]["mean"] == pytest.approx(1.0)
        assert stats["spreads"]["spread_oh"]["median"] == pytest.approx(1.0)


class TestSpreadLabels:
    def test_all_keys_present(self):
        expected_keys = {
            "spread_oh", "spread_ol", "spread_hl",
            "spread_oc", "spread_hc", "spread_lc",
        }
        assert set(SPREAD_LABELS.keys()) == expected_keys

    def test_labels_are_chinese(self):
        for label in SPREAD_LABELS.values():
            assert any("\u4e00" <= c <= "\u9fff" for c in label)
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /root/code/github/travelliu/stock && python -m pytest tests/test_analysis.py -v`
Expected: FAIL — `ModuleNotFoundError: No module named 'analysis'`

- [ ] **Step 3: Write minimal implementation**

```python
# analysis.py
import statistics

from typing import Any

SPREAD_KEYS = [
    "spread_oh", "spread_ol", "spread_hl",
    "spread_oc", "spread_hc", "spread_lc",
]

SPREAD_LABELS: dict[str, str] = {
    "spread_oh": "最高-开盘",
    "spread_ol": "开盘-最低",
    "spread_hl": "最高-最低",
    "spread_oc": "开盘-收盘",
    "spread_hc": "最高-收盘",
    "spread_lc": "最低-收盘",
}


def compute_statistics(
    rows: list[dict[str, Any]],
) -> dict[str, Any]:
    """Calculate mean and median for each spread type.

    Args:
        rows: list of row dicts from db.query_daily()

    Returns:
        {"count": N, "spreads": {"spread_oh": {"mean": X, "median": Y}, ...}}
    """
    if not rows:
        return {"count": 0, "spreads": {}}

    spreads: dict[str, dict[str, float]] = {}
    for key in SPREAD_KEYS:
        values = [r[key] for r in rows if r.get(key) is not None]
        if values:
            spreads[key] = {
                "mean": statistics.mean(values),
                "median": statistics.median(values),
            }

    return {"count": len(rows), "spreads": spreads}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /root/code/github/travelliu/stock && python -m pytest tests/test_analysis.py -v`
Expected: all 5 tests PASS

- [ ] **Step 5: Commit**

```bash
git add analysis.py tests/test_analysis.py
git commit -m "feat: add analysis module with spread statistics"
```

---

## Task 5: stock.py — CLI Entry Point

**Files:**
- Create: `stock.py`
- Create: `tests/test_cli.py`

- [ ] **Step 1: Write the failing test**

```python
# tests/test_cli.py
import subprocess
import sys


def run_cli(*args: str) -> subprocess.CompletedProcess:
    return subprocess.run(
        [sys.executable, "stock.py", *args],
        capture_output=True,
        text=True,
        cwd="/root/code/github/travelliu/stock",
    )


class TestCLIHelp:
    def test_no_args_shows_help(self):
        result = run_cli()
        assert result.returncode != 0 or "usage" in result.stdout.lower() or "usage" in result.stderr.lower()

    def test_fetch_help(self):
        result = run_cli("fetch", "--help")
        assert result.returncode == 0
        assert "--stocks" in result.stdout
        assert "--days" in result.stdout

    def test_show_help(self):
        result = run_cli("show", "--help")
        assert result.returncode == 0
        assert "--stock" in result.stdout
        assert "--analyze" in result.stdout


class TestCLIShow:
    def test_show_requires_stock(self):
        result = run_cli("show")
        assert result.returncode != 0
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /root/code/github/travelliu/stock && python -m pytest tests/test_cli.py -v`
Expected: FAIL — `FileNotFoundError` or module not found

- [ ] **Step 3: Write minimal implementation**

```python
#!/usr/bin/env python3
# stock.py — CLI entry point

import argparse
import sys
from datetime import datetime, timedelta

from tabulate import tabulate

from config import DEFAULT_STOCKS, DEFAULT_FETCH_DAYS, DB_PATH
from db import DailyDB
from fetcher import fetch_daily
from analysis import compute_statistics, SPREAD_LABELS, SPREAD_KEYS


def cmd_fetch(args: argparse.Namespace) -> None:
    stocks = args.stocks.split(",") if args.stocks else DEFAULT_STOCKS
    days = args.days or DEFAULT_FETCH_DAYS
    db = DailyDB(DB_PATH)
    db.init()

    for code in stocks:
        print(f"Fetching {code} ...")
        df = fetch_daily(code, days=days)
        if df.empty:
            print(f"  No data returned for {code}")
            continue
        db.insert_daily(df)
        print(f"  Inserted {len(df)} rows")


def cmd_show(args: argparse.Namespace) -> None:
    stock = args.stock
    end_date = args.to or datetime.now().strftime("%Y-%m-%d")
    start_date = args.from_ or (
        datetime.now() - timedelta(days=30)
    ).strftime("%Y-%m-%d")

    db = DailyDB(DB_PATH)
    rows = db.query_daily(stock, start_date, end_date)

    if not rows:
        print(f"No data for {stock} ({start_date} ~ {end_date})")
        return

    if args.analyze:
        _print_analysis(stock, start_date, end_date, rows)
    else:
        _print_table(rows)


def _print_table(rows: list[dict]) -> None:
    headers = [
        "trade_date", "open", "high", "low", "close",
        "vol", "spread_oh", "spread_ol", "spread_hl",
        "spread_oc", "spread_hc", "spread_lc",
    ]
    display_names = [
        "日期", "开盘", "最高", "最低", "收盘",
        "成交量", "高-开", "开-低", "高-低",
        "开-收", "高-收", "低-收",
    ]
    table = [[r.get(h, "") for h in headers] for r in rows]
    print(tabulate(table, headers=display_names, floatfmt=".2f"))


def _print_analysis(
    stock: str, start_date: str, end_date: str, rows: list[dict]
) -> None:
    stats = compute_statistics(rows)
    print(f"=== {stock} 价差分析 ({start_date} ~ {end_date}) ===")
    print(f"样本数: {stats['count']}")
    print()
    print(f"{'价差类型':<14} {'平均值':>8} {'中位数':>8}")
    print("-" * 32)
    for key in SPREAD_KEYS:
        if key in stats["spreads"]:
            label = SPREAD_LABELS[key]
            s = stats["spreads"][key]
            print(f"{label:<14} {s['mean']:>8.2f} {s['median']:>8.2f}")


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Stock Daily Data CLI (Tushare)"
    )
    sub = parser.add_subparsers(dest="command")

    # fetch
    p_fetch = sub.add_parser("fetch", help="Fetch daily data from tushare")
    p_fetch.add_argument("--stocks", type=str, default=None,
                         help="Comma-separated stock codes (default: all 8)")
    p_fetch.add_argument("--days", type=int, default=None,
                         help=f"Lookback days (default: {DEFAULT_FETCH_DAYS})")

    # show
    p_show = sub.add_parser("show", help="Show daily data and analysis")
    p_show.add_argument("--stock", type=str, required=True,
                        help="Stock code (e.g., 603778)")
    p_show.add_argument("--from", dest="from_", type=str, default=None,
                        help="Start date (YYYY-MM-DD, default: 30 days ago)")
    p_show.add_argument("--to", type=str, default=None,
                        help="End date (YYYY-MM-DD, default: today)")
    p_show.add_argument("--analyze", action="store_true",
                        help="Show spread statistics")

    return parser


def main() -> None:
    parser = build_parser()
    args = parser.parse_args()
    if args.command == "fetch":
        cmd_fetch(args)
    elif args.command == "show":
        cmd_show(args)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /root/code/github/travelliu/stock && python -m pytest tests/test_cli.py -v`
Expected: all 3 tests PASS

- [ ] **Step 5: Commit**

```bash
git add stock.py tests/test_cli.py
git commit -m "feat: add CLI entry point with fetch and show subcommands"
```

---

## Task 6: Update requirements.txt

**Files:**
- Modify: `requirements.txt`

- [ ] **Step 1: Update requirements.txt**

Replace `akshare>=1.14.0` with `tushare>=1.4.0`:

```
tushare>=1.4.0
pandas>=2.0.0
tabulate>=0.9.0
pytest>=7.0.0
```

- [ ] **Step 2: Install new dependencies**

Run: `cd /root/code/github/travelliu/stock && pip install -r requirements.txt`

- [ ] **Step 3: Commit**

```bash
git add requirements.txt
git commit -m "chore: replace akshare with tushare in requirements"
```

---

## Task 7: Integration Test with 603778

**Files:**
- Create: `tests/test_integration.py`

This task verifies the full pipeline against the real tushare API using stock **603778**.

- [ ] **Step 1: Write integration test**

```python
# tests/test_integration.py
"""Integration test: fetch real data for 603778 and verify end-to-end flow."""
import subprocess
import sys

import pytest

from config import DB_PATH
from db import DailyDB
from fetcher import fetch_daily


@pytest.fixture(autouse=True)
def clean_db():
    """Remove test DB before and after integration test."""
    if DB_PATH.exists():
        DB_PATH.unlink()
    yield
    if DB_PATH.exists():
        DB_PATH.unlink()


class TestIntegration603778:
    def test_fetch_and_store(self):
        df = fetch_daily("603778", days=30)
        assert not df.empty
        assert "spread_oh" in df.columns
        assert "trade_date" in df.columns
        # Date format should be YYYY-MM-DD
        assert "-" in df.iloc[0]["trade_date"]

        db = DailyDB(DB_PATH)
        db.init()
        count = db.insert_daily(df)
        assert count > 0

        rows = db.query_daily("603778", "2020-01-01", "2099-12-31")
        assert len(rows) > 0
        assert rows[0]["ts_code"] == "603778.SH"

    def test_cli_fetch_show(self):
        # Fetch via CLI
        result = subprocess.run(
            [sys.executable, "stock.py", "fetch", "--stocks", "603778", "--days", "30"],
            capture_output=True, text=True,
            cwd="/root/code/github/travelliu/stock",
        )
        assert result.returncode == 0
        assert "603778" in result.stdout

        # Show via CLI
        result = subprocess.run(
            [sys.executable, "stock.py", "show", "--stock", "603778"],
            capture_output=True, text=True,
            cwd="/root/code/github/travelliu/stock",
        )
        assert result.returncode == 0
        assert "603778" in result.stdout or "SH" in result.stdout

        # Analyze via CLI
        result = subprocess.run(
            [sys.executable, "stock.py", "show", "--stock", "603778", "--analyze"],
            capture_output=True, text=True,
            cwd="/root/code/github/travelliu/stock",
        )
        assert result.returncode == 0
        assert "价差分析" in result.stdout
        assert "样本数" in result.stdout
```

- [ ] **Step 2: Run integration test**

Run: `cd /root/code/github/travelliu/stock && python -m pytest tests/test_integration.py -v --timeout=60`

Expected: all 2 tests PASS (requires network access to tushare API)

- [ ] **Step 3: Commit**

```bash
git add tests/test_integration.py
git commit -m "test: add integration test for 603778 end-to-end flow"
```

---

## Task 8: Run Full Test Suite & Verify

- [ ] **Step 1: Run all tests**

Run: `cd /root/code/github/travelliu/stock && python -m pytest tests/ -v --tb=short`

Expected: all tests PASS

- [ ] **Step 2: Verify CLI manually with 603778**

```bash
cd /root/code/github/travelliu/stock
python stock.py fetch --stocks 603778 --days 30
python stock.py show --stock 603778
python stock.py show --stock 603778 --analyze
```

Expected: each command runs without error, `show --analyze` outputs formatted spread statistics table.

- [ ] **Step 3: Clean up deleted files from git tracking**

```bash
git rm init_data.py morning_scan.py evening_close.py analyzer.py 2>/dev/null || true
git rm tests/test_analyzer.py tests/test_config.py tests/test_integration.py tests/test_morning_scan.py 2>/dev/null || true
git rm .gitattributes 2>/dev/null || true
```

Note: only remove files that the design spec lists as deleted AND that still exist in git index.

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat: complete tushare daily data system with CLI, DB, analysis"
```
