# analyzer.py
"""Analyze auction volume ratios and output sorted report."""

from datetime import date, timedelta
from pathlib import Path

import pandas as pd
from tabulate import tabulate


def load_auction_history(auction_dir: Path, today: date, days: int = 5):
    """Load auction CSV files from the past N days including today."""
    frames = []
    for i in range(days):
        d = today - timedelta(days=i)
        csv_path = auction_dir / f"{d}.csv"
        if csv_path.exists():
            df = pd.read_csv(csv_path, dtype={"股票代码": str})
            df["日期"] = d.strftime("%Y-%m-%d")
            frames.append(df)
    if not frames:
        return pd.DataFrame()
    return pd.concat(frames, ignore_index=True)


def load_daily_history(daily_dir: Path, symbol: str, days: int = 5):
    """Load daily OHLCV CSV for a single stock, return last N rows."""
    csv_path = daily_dir / f"{symbol}.csv"
    if not csv_path.exists():
        return pd.DataFrame()
    df = pd.read_csv(csv_path, dtype={"成交量": float})
    return df.tail(days)


def calc_volume_ratio_a(today_auction_vol: float, hist_auction_vols: pd.Series):
    """竞价量比 A = 今日竞价量 / 过去 N 日平均竞价量."""
    if hist_auction_vols.empty or hist_auction_vols.mean() == 0:
        return float("nan")
    return today_auction_vol / hist_auction_vols.mean()


def calc_volume_ratio_b(today_auction_vol: float, daily_df: pd.DataFrame):
    """竞价量比 B = 今日竞价量 / 过去 N 日平均全天成交量."""
    if daily_df.empty or daily_df["成交量"].mean() == 0:
        return float("nan")
    return today_auction_vol / daily_df["成交量"].mean()


def build_report(
    today_auction: pd.DataFrame,
    auction_history: pd.DataFrame,
    daily_dir: Path,
    today: date,
    history_days: int = 5,
) -> pd.DataFrame:
    """Build the full report DataFrame with volume ratios."""
    today_str = today.strftime("%Y-%m-%d")
    rows = []

    for _, row in today_auction.iterrows():
        symbol = row["股票代码"]
        name = row["股票名称"]
        auction_vol = float(row["竞价成交量"])
        open_price = row.get("竞价成交价", float("nan"))

        hist = auction_history[
            (auction_history["股票代码"] == symbol) & (auction_history["日期"] != today_str)
        ]
        ratio_a = calc_volume_ratio_a(auction_vol, hist["竞价成交量"])

        daily_df = load_daily_history(daily_dir, symbol, days=history_days)
        ratio_b = calc_volume_ratio_b(auction_vol, daily_df)

        if not daily_df.empty and open_price and open_price == open_price:
            last_close = daily_df.iloc[-1]["收盘"]
            open_change = (open_price - last_close) / last_close * 100 if last_close else float("nan")
        else:
            open_change = float("nan")

        rows.append({
            "代码": symbol,
            "名称": name,
            "今日竞价量": auction_vol,
            "竞价量比A": ratio_a,
            "竞价量比B": ratio_b,
            "开盘价": open_price,
            "开盘涨跌幅": round(open_change, 2) if open_change == open_change else "N/A",
        })

    report = pd.DataFrame(rows)
    report = report.sort_values("竞价量比A", ascending=False, na_position="last").reset_index(drop=True)
    return report


def format_report(report: pd.DataFrame):
    """Pretty-print the report table to terminal."""
    if report.empty:
        print("No auction data to display.")
        return

    display = report.copy()
    display["今日竞价量"] = display["今日竞价量"].apply(lambda x: f"{x:,.0f}")
    display["竞价量比A"] = display["竞价量比A"].apply(
        lambda x: f"{x:.2f}" if x == x else "N/A"
    )
    display["竞价量比B"] = display["竞价量比B"].apply(
        lambda x: f"{x:.2f}" if x == x else "N/A"
    )
    display["开盘价"] = display["开盘价"].apply(
        lambda x: f"{x:.2f}" if x == x else "N/A"
    )

    table = tabulate(display, headers="keys", tablefmt="grid", showindex=False)
    print(table)
