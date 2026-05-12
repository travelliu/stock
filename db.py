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

    def get_max_date(self, stock_code: str) -> str | None:
        """Return the latest trade_date for a stock, or None if no data."""
        from config import to_tushare_code

        ts_code = to_tushare_code(stock_code)
        conn = sqlite3.connect(self.db_path)
        cursor = conn.execute(
            "SELECT MAX(trade_date) FROM daily WHERE ts_code = ?",
            (ts_code,),
        )
        row = cursor.fetchone()
        conn.close()
        return row[0] if row and row[0] else None

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
