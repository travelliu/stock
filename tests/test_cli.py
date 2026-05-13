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
        assert "--all" in result.stdout


class TestCLIShow:
    def test_show_requires_stock(self):
        result = run_cli("show")
        assert result.returncode != 0


class TestRecommendationInOutput:
    def test_default_output_contains_recommendation(self, capsys):
        from analysis import StockAnalyzer

        analyzer = StockAnalyzer("603778")
        analyzer.all_rows = [
            {"trade_date": "2026-05-12", "open": 13.58, "high": 13.86,
             "low": 12.87, "close": 13.38, "vol": 742300,
             "spread_oh": 0.28, "spread_ol": 0.71, "spread_hl": 0.99,
             "spread_oc": 0.20, "spread_hc": 0.48, "spread_lc": 0.51},
            {"trade_date": "2026-05-11", "open": 13.50, "high": 13.70,
             "low": 13.20, "close": 13.60, "vol": 500000,
             "spread_oh": 0.20, "spread_ol": 0.30, "spread_hl": 0.50,
             "spread_oc": 0.10, "spread_hc": 0.10, "spread_lc": 0.40},
            {"trade_date": "2026-05-08", "open": 13.40, "high": 13.90,
             "low": 13.10, "close": 13.50, "vol": 600000,
             "spread_oh": 0.50, "spread_ol": 0.30, "spread_hl": 0.80,
             "spread_oc": 0.10, "spread_hc": 0.40, "spread_lc": 0.40},
        ]
        analyzer.print_analysis()
        captured = capsys.readouterr()
        assert "高抛低吸推荐" in captured.out
        assert "累计占比" in captured.out

    def test_show_all_output_excludes_recommendation(self, capsys):
        from analysis import StockAnalyzer

        analyzer = StockAnalyzer("603778", show_all=True)
        analyzer.all_rows = [
            {"trade_date": "2026-05-12", "open": 13.58, "high": 13.86,
             "low": 12.87, "close": 13.38, "vol": 742300,
             "spread_oh": 0.28, "spread_ol": 0.71, "spread_hl": 0.99,
             "spread_oc": 0.20, "spread_hc": 0.48, "spread_lc": 0.51},
        ]
        analyzer.print_analysis()
        captured = capsys.readouterr()
        assert "高抛低吸推荐" not in captured.out


import pytest


class TestTradingPlanOption:
    def test_parser_accepts_open(self):
        from stock import build_parser
        parser = build_parser()
        args = parser.parse_args(["show", "--stock", "603778", "--open", "200"])
        assert args.open == 200.0

    def test_parser_rejects_invalid_open(self):
        from stock import build_parser
        parser = build_parser()
        with pytest.raises(SystemExit):
            parser.parse_args(["show", "--stock", "603778", "--open", "abc"])

    def test_parser_accepts_low(self):
        from stock import build_parser
        parser = build_parser()
        args = parser.parse_args(["show", "--stock", "603778", "--low", "13.0"])
        assert args.low == 13.0

    def test_parser_accepts_high(self):
        from stock import build_parser
        parser = build_parser()
        args = parser.parse_args(["show", "--stock", "603778", "--high", "14.0"])
        assert args.high == 14.0

    def test_show_without_open_no_plan(self, capsys, monkeypatch):
        from stock import build_parser, cmd_show
        import db, company
        parser = build_parser()
        args = parser.parse_args(["show", "--stock", "603778"])
        class FakeDB:
            def init(self): pass
            def query_daily(self, *a, **k): return []
            def get_max_date(self, *a, **k): return None
        monkeypatch.setattr(db, "DailyDB", lambda path: FakeDB())
        monkeypatch.setattr(company, "get_stock_name", lambda code: "")
        cmd_show(args)
        captured = capsys.readouterr()
        assert "交易计划" not in captured.out
