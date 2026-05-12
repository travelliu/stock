"""Integration test: fetch real data for 603778 and verify end-to-end flow."""
import pytest

from config import DB_PATH
from db import DailyDB
from fetcher import fetch_daily


@pytest.fixture(autouse=True)
def use_test_db(tmp_path, monkeypatch):
    """Redirect DB_PATH to a temp file so real data is never touched."""
    test_db = tmp_path / "test.db"
    monkeypatch.setattr("config.DB_PATH", test_db)
    monkeypatch.setattr("db.DB_PATH", test_db, raising=False)
    yield


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
        assert isinstance(df2, type(df1))

    def test_cli_fetch_show(self, capsys):
        from stock import cmd_fetch, cmd_show, build_parser

        # Simulate: python stock.py fetch --stocks 603778 --days 30
        args = build_parser().parse_args(["fetch", "--stocks", "603778", "--days", "30"])
        cmd_fetch(args)
        out = capsys.readouterr().out
        assert "603778" in out
        assert "Inserted" in out

        # Simulate: python stock.py show --stock 603778
        args = build_parser().parse_args(["show", "--stock", "603778"])
        cmd_show(args)
        out = capsys.readouterr().out
        assert "价差分析" in out
