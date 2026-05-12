import sqlite3
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


class TestGetMaxDate:
    def test_returns_none_when_empty(self, db):
        assert db.get_max_date("603778") is None

    def test_returns_max_date(self, db):
        df = pd.DataFrame([
            {"ts_code": "603778.SH", "trade_date": "2025-05-10",
             "open": 10.0, "high": 11.0, "low": 9.5, "close": 10.5,
             "vol": 1000.0, "amount": 10000.0,
             "spread_oh": 1.0, "spread_ol": 0.5, "spread_hl": 1.5,
             "spread_oc": -0.5, "spread_hc": 0.5, "spread_lc": -1.0},
            {"ts_code": "603778.SH", "trade_date": "2025-05-15",
             "open": 12.0, "high": 13.0, "low": 11.5, "close": 12.5,
             "vol": 3000.0, "amount": 30000.0,
             "spread_oh": 1.0, "spread_ol": 0.5, "spread_hl": 1.5,
             "spread_oc": -0.5, "spread_hc": 0.5, "spread_lc": -1.0},
        ])
        db.insert_daily(df)
        assert db.get_max_date("603778") == "2025-05-15"

    def test_returns_none_for_other_stock(self, db, sample_df):
        db.insert_daily(sample_df)
        assert db.get_max_date("000890") is None
