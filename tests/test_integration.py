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


@pytest.mark.integration
class TestIntegration603778:
    def test_fetch_and_store(self):
        df = fetch_daily("603778", days=30)
        assert not df.empty
        assert "spread_oh" in df.columns
        assert "trade_date" in df.columns
        assert "-" in df.iloc[0]["trade_date"]

        db = DailyDB(DB_PATH)
        db.init()
        count = db.insert_daily(df)
        assert count > 0

        rows = db.query_daily("603778", "2020-01-01", "2099-12-31")
        assert len(rows) > 0
        assert rows[0]["ts_code"] == "603778.SH"

    def test_incremental_fetch(self):
        db = DailyDB(DB_PATH)
        db.init()

        # First fetch
        df1 = fetch_daily("603778", days=30)
        assert not df1.empty
        db.insert_daily(df1)

        max_date = db.get_max_date("603778")
        assert max_date is not None

        # Incremental fetch should use max_date as start
        df2 = fetch_daily("603778", start_date=max_date)
        # May or may not have new data depending on market hours
        # But the call should not error
        assert isinstance(df2, type(df1))

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

        # Analyze via CLI
        result = subprocess.run(
            [sys.executable, "stock.py", "show", "--stock", "603778", "--analyze"],
            capture_output=True, text=True,
            cwd="/root/code/github/travelliu/stock",
        )
        assert result.returncode == 0
        assert "价差分析" in result.stdout
        assert "样本数" in result.stdout
