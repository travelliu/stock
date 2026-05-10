# tests/test_integration.py
"""End-to-end integration test using mock data."""

from datetime import date, timedelta
from pathlib import Path

import pandas as pd
import pytest
import sys

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from analyzer import build_report, load_auction_history


@pytest.fixture
def mock_project(tmp_path):
    daily_dir = tmp_path / "daily"
    daily_dir.mkdir()
    auction_dir = tmp_path / "auction"
    auction_dir.mkdir()

    symbols = ["000001", "000002", "600000"]
    names = ["平安银行", "万科A", "浦发银行"]
    auction_base = [200000, 150000, 300000]

    today = date(2026, 5, 10)

    for i in range(5, 0, -1):
        d = today - timedelta(days=i)
        rows = []
        for s, n, av in zip(symbols, names, auction_base):
            rows.append({
                "股票代码": s, "股票名称": n,
                "竞价成交价": 10.0 + i * 0.5,
                "竞价成交量": av + i * 10000,
                "竞价成交额": (av + i * 10000) * (10.0 + i * 0.5),
                "未匹配量": 50000, "粗筛标记": True,
            })
        pd.DataFrame(rows).to_csv(auction_dir / f"{d}.csv", index=False)

    today_rows = [
        {"股票代码": "000001", "股票名称": "平安银行",
         "竞价成交价": 15.0, "竞价成交量": 2000000,
         "竞价成交额": 30000000, "未匹配量": 500000, "粗筛标记": True},
        {"股票代码": "000002", "股票名称": "万科A",
         "竞价成交价": 8.0, "竞价成交量": 160000,
         "竞价成交额": 1280000, "未匹配量": 20000, "粗筛标记": True},
        {"股票代码": "600000", "股票名称": "浦发银行",
         "竞价成交价": 7.5, "竞价成交量": 310000,
         "竞价成交额": 2325000, "未匹配量": 30000, "粗筛标记": True},
    ]
    pd.DataFrame(today_rows).to_csv(auction_dir / f"{today}.csv", index=False)

    for s in symbols:
        rows = [
            {"日期": f"2026-05-{d:02d}", "开盘": 10.0, "收盘": 10.5, "最高": 10.8,
             "最低": 9.9, "成交量": 5000000, "成交额": 52500000,
             "振幅": 8.5, "涨跌幅": 1.5, "涨跌额": 0.15, "换手率": 1.2}
            for d in range(5, 10)
        ]
        pd.DataFrame(rows).to_csv(daily_dir / f"{s}.csv", index=False)

    return {"daily_dir": daily_dir, "auction_dir": auction_dir, "today": today}


def test_full_pipeline(mock_project):
    auction_dir = mock_project["auction_dir"]
    daily_dir = mock_project["daily_dir"]
    today = mock_project["today"]

    auction_history = load_auction_history(auction_dir, today, days=5)
    today_str = today.strftime("%Y-%m-%d")
    today_auction = auction_history[auction_history["日期"] == today_str]

    report = build_report(today_auction, auction_history, daily_dir, today)

    assert report.iloc[0]["代码"] == "000001"
    assert report.iloc[0]["竞价量比A"] > 5.0
    assert len(report) == 3

    ratios = report["竞价量比A"].tolist()
    assert ratios == sorted(ratios, reverse=True)


def test_cold_start(mock_project):
    auction_dir = mock_project["auction_dir"]
    daily_dir = mock_project["daily_dir"]
    today = mock_project["today"]

    for f in auction_dir.iterdir():
        if today.strftime("%Y-%m-%d") not in f.name:
            f.unlink()

    auction_history = load_auction_history(auction_dir, today, days=5)
    today_str = today.strftime("%Y-%m-%d")
    today_auction = auction_history[auction_history["日期"] == today_str]

    report = build_report(today_auction, auction_history, daily_dir, today)

    assert len(report) == 3
    assert report.iloc[0]["竞价量比B"] > 0
