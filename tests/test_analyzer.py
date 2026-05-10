# tests/test_analyzer.py
import sys
from datetime import date
from pathlib import Path

import pandas as pd
import pytest

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from analyzer import (
    load_auction_history,
    load_daily_history,
    calc_volume_ratio_a,
    calc_volume_ratio_b,
    build_report,
    format_report,
)


@pytest.fixture
def tmp_data_dirs(tmp_path):
    auction_dir = tmp_path / "auction"
    auction_dir.mkdir()
    daily_dir = tmp_path / "daily"
    daily_dir.mkdir()

    for i in range(5, 0, -1):
        d = date(2026, 5, 10 - i)
        df = pd.DataFrame([
            {
                "股票代码": "000001",
                "股票名称": "平安银行",
                "竞价成交价": 12.0 + i * 0.1,
                "竞价成交量": 1000000 + i * 100000,
                "竞价成交额": 12000000 + i * 1200000,
                "未匹配量": 500000,
                "粗筛标记": True,
            }
        ])
        df.to_csv(auction_dir / f"{d}.csv", index=False)

    today = date(2026, 5, 10)
    today_auction = pd.DataFrame([
        {
            "股票代码": "000001",
            "股票名称": "平安银行",
            "竞价成交价": 13.0,
            "竞价成交量": 5000000,
            "竞价成交额": 65000000,
            "未匹配量": 2000000,
            "粗筛标记": True,
        }
    ])
    today_auction.to_csv(auction_dir / f"{today}.csv", index=False)

    daily_data = pd.DataFrame([
        {
            "日期": f"2026-05-{d:02d}", "开盘": 12.5, "收盘": 12.8, "最高": 13.0,
            "最低": 12.4, "成交量": 10000000, "成交额": 128000000,
            "振幅": 4.0, "涨跌幅": 2.4, "涨跌额": 0.3, "换手率": 2.5,
        }
        for d in range(5, 10)
    ])
    daily_data.to_csv(daily_dir / "000001.csv", index=False)

    return auction_dir, daily_dir


def test_load_auction_history(tmp_data_dirs):
    auction_dir, _ = tmp_data_dirs
    today = date(2026, 5, 10)
    history = load_auction_history(auction_dir, today, days=5)
    assert len(history) == 5  # 5 days including today
    assert "竞价成交量" in history.columns


def test_load_daily_history(tmp_data_dirs):
    _, daily_dir = tmp_data_dirs
    df = load_daily_history(daily_dir, "000001", days=5)
    assert len(df) == 5
    assert "成交量" in df.columns


def test_calc_volume_ratio_a(tmp_data_dirs):
    auction_dir, _ = tmp_data_dirs
    today = date(2026, 5, 10)
    history = load_auction_history(auction_dir, today, days=5)
    today_str = today.strftime("%Y-%m-%d")
    today_vol = history[history["日期"] == today_str]["竞价成交量"].iloc[0]
    hist_vols = history[history["日期"] != today_str]["竞价成交量"]
    ratio = calc_volume_ratio_a(today_vol, hist_vols)
    assert ratio > 1.0


def test_calc_volume_ratio_b(tmp_data_dirs):
    _, daily_dir = tmp_data_dirs
    today_vol = 5000000
    daily_df = load_daily_history(daily_dir, "000001", days=5)
    ratio = calc_volume_ratio_b(today_vol, daily_df)
    assert ratio > 0


def test_build_report(tmp_data_dirs):
    auction_dir, daily_dir = tmp_data_dirs
    today = date(2026, 5, 10)
    auction_history = load_auction_history(auction_dir, today, days=5)
    today_str = today.strftime("%Y-%m-%d")
    today_auction = auction_history[auction_history["日期"] == today_str]

    report = build_report(
        today_auction=today_auction,
        auction_history=auction_history,
        daily_dir=daily_dir,
        today=today,
    )
    assert len(report) > 0
    assert "代码" in report.columns
    assert "竞价量比A" in report.columns
    assert "竞价量比B" in report.columns


def test_format_report(tmp_data_dirs, capsys):
    auction_dir, daily_dir = tmp_data_dirs
    today = date(2026, 5, 10)
    auction_history = load_auction_history(auction_dir, today, days=5)
    today_str = today.strftime("%Y-%m-%d")
    today_auction = auction_history[auction_history["日期"] == today_str]
    report = build_report(
        today_auction=today_auction,
        auction_history=auction_history,
        daily_dir=daily_dir,
        today=today,
    )
    format_report(report)
    captured = capsys.readouterr()
    assert "000001" in captured.out
    assert "平安银行" in captured.out
