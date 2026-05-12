#!/usr/bin/env python3
import argparse
from datetime import datetime, timedelta

from tabulate import tabulate

from config import DEFAULT_FETCH_DAYS, DB_PATH
from db import DailyDB
from fetcher import fetch_daily
from analysis import compute_statistics, SPREAD_LABELS, SPREAD_KEYS


def cmd_fetch(args: argparse.Namespace) -> None:
    stocks = args.stocks.split(",")
    db = DailyDB(DB_PATH)
    db.init()

    for code in stocks:
        code = code.strip()
        print(f"Fetching {code} ...")

        if args.days == "all":
            # Full fetch: lookback 10 years
            days = 3650
            start_date = None
        elif args.days:
            days = int(args.days)
            start_date = None
        else:
            # Default: incremental — fetch from last known date
            max_date = db.get_max_date(code)
            if max_date:
                start_date = max_date
                days = None
                print(f"  Last record: {max_date}, fetching incremental ...")
            else:
                days = DEFAULT_FETCH_DAYS
                start_date = None
                print(f"  No existing data, fetching {days} days ...")

        df = fetch_daily(code, start_date=start_date, days=days)
        if df.empty:
            print(f"  No new data for {code}")
            continue
        count = db.insert_daily(df)
        print(f"  Inserted {count} rows")


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
        print(f"--- {stock} 日线数据 ({start_date} ~ {end_date}) 共 {len(rows)} 条 ---")
        print()
        _print_table(rows)


def _print_table(rows: list[dict]) -> None:
    headers = [
        "trade_date", "open", "high", "low", "close",
        "vol", "spread_oh", "spread_ol", "spread_hl",
        "spread_oc", "spread_hc", "spread_lc",
    ]
    display_names = [
        "日期", "开盘", "最高", "最低", "收盘",
        "成交量(万)", "高-开", "开-低", "高-低",
        "开-收", "高-收", "低-收",
    ]
    # Reverse: newest first
    table = [[r.get(h, "") for h in headers] for r in reversed(rows)]
    # Format volume to 万手
    for row in table:
        vol_idx = headers.index("vol")
        row[vol_idx] = round(row[vol_idx] / 10000, 2)
    print(tabulate(table, headers=display_names, floatfmt=".2f",
                   tablefmt="rounded_outline", stralign="right"))


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
    p_fetch.add_argument("--stocks", type=str, required=True,
                         help="Comma-separated stock codes (e.g., 603778,000890)")
    p_fetch.add_argument("--days", type=str, default=None,
                         help="Lookback days (e.g., 30, 180) or 'all' for full history. "
                              "Default: incremental from last record")

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
