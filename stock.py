#!/usr/bin/env python3
import argparse
import statistics
import unicodedata
from datetime import datetime, timedelta

from config import DEFAULT_FETCH_DAYS, DB_PATH
from db import DailyDB
from fetcher import fetch_daily
from analysis import (
    compute_statistics, compute_distribution,
    SPREAD_LABELS, SPREAD_KEYS, DEFAULT_SPREADS,
)


def _display_width(s: str) -> int:
    """Calculate terminal display width (CJK chars = 2 columns)."""
    width = 0
    for ch in str(s):
        eaw = unicodedata.east_asian_width(ch)
        width += 2 if eaw in ("W", "F") else 1
    return width


def _rpad(s: str, width: int) -> str:
    """Right-pad string to reach target display width."""
    return str(s) + " " * max(0, width - _display_width(s))


def _lpad(s: str, width: int) -> str:
    """Left-pad string to reach target display width."""
    return " " * max(0, width - _display_width(s)) + str(s)


def _format_table(headers: list[str], rows: list[list[str]]) -> str:
    """Format a table with CJK-aware column alignment."""
    # Calculate column widths
    col_widths = [_display_width(h) for h in headers]
    for row in rows:
        for i, cell in enumerate(row):
            col_widths[i] = max(col_widths[i], _display_width(cell))

    # Build horizontal separator
    sep = "+" + "+".join("-" * (w + 2) for w in col_widths) + "+"

    lines = [sep]
    # Header row
    header_line = "|"
    for i, h in enumerate(headers):
        header_line += " " + _rpad(h, col_widths[i]) + " |"
    lines.append(header_line)
    lines.append(sep)

    # Data rows
    for row in rows:
        data_line = "|"
        for i, cell in enumerate(row):
            data_line += " " + _rpad(cell, col_widths[i]) + " |"
        lines.append(data_line)
    lines.append(sep)

    return "\n".join(lines)


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
    db = DailyDB(DB_PATH)
    db.init()

    end_date = args.to or datetime.now().strftime("%Y-%m-%d")
    start_date = args.from_ or (
        datetime.now() - timedelta(days=30)
    ).strftime("%Y-%m-%d")
    show_all = getattr(args, "all", False)

    # Always show analysis (uses all data for multi-window)
    all_rows = db.query_daily(stock, "2000-01-01", "2099-12-31")
    if all_rows:
        _print_analysis(stock, all_rows, show_all=show_all)

    # Also show raw data table for the date range
    rows = db.query_daily(stock, start_date, end_date)
    if rows:
        print(f"--- {stock} 日线数据 ({start_date} ~ {end_date}) 共 {len(rows)} 条 ---")
        print()
        _print_table(rows)
    elif not all_rows:
        print(f"No data for {stock}")


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
    # Reverse: newest first, format values
    table = []
    for r in reversed(rows):
        row = []
        for h in headers:
            v = r.get(h, "")
            if h == "vol":
                row.append(f"{v / 10000:>8.2f}")
            elif isinstance(v, float):
                row.append(f"{v:>6.2f}")
            else:
                row.append(str(v))
        table.append(row)

    print(_format_table(display_names, table))


def _join_tables_side_by_side(tables: list[str], gaps: int = 4) -> str:
    """Join multiple table strings side by side with CJK-aware alignment."""
    if not tables:
        return ""
    split = [t.split("\n") for t in tables]

    # Normalize each block: pad all lines to the max display width of that block
    normalized = []
    for block in split:
        if not block:
            continue
        max_w = max(_display_width(line) for line in block)
        padded = [_rpad(line, max_w) for line in block]
        normalized.append(padded)

    max_lines = max(len(b) for b in normalized)
    pad = " " * gaps
    lines = []
    for i in range(max_lines):
        parts = []
        for b in normalized:
            parts.append(b[i] if i < len(b) else " " * _display_width(b[0]))
        lines.append(pad.join(parts))
    return "\n".join(lines)


def _print_analysis(
    stock: str, all_rows: list[dict], show_all: bool = False
) -> None:
    # Build time windows by trading days (newest first in each window)
    all_rows_sorted = sorted(all_rows, key=lambda r: r["trade_date"], reverse=True)
    windows = [
        ("全部", all_rows_sorted),
        ("近90日", all_rows_sorted[:90]),
        ("近30日", all_rows_sorted[:30]),
        ("近15日", all_rows_sorted[:15]),
    ]

    spread_keys = SPREAD_KEYS if show_all else DEFAULT_SPREADS

    print(f"=== {stock} 价差分析 ===")
    print()

    for key in spread_keys:
        label = SPREAD_LABELS[key]

        # --- Summary table ---
        headers = ["时段", "样本数", "均值", "中位数"]
        table = []
        for name, rows in windows:
            values = [r[key] for r in rows if r.get(key) is not None]
            if values:
                table.append([
                    name,
                    str(len(values)),
                    f"{statistics.mean(values):.2f}",
                    f"{statistics.median(values):.2f}",
                ])
            else:
                table.append([name, "0", "-", "-"])

        summary_str = f"── {label} 汇总 ──\n" + _format_table(headers, table)

        # --- Distribution tables side by side ---
        dist_tables = []
        for name, rows in windows:
            values = [r[key] for r in rows if r.get(key) is not None]
            if not values:
                continue
            bins = compute_distribution(values)
            dist_headers = ["区间", "数量", "占比"]
            dist_table = []
            for b in bins:
                interval = f"{b['low']:.2f}~{b['high']:.2f}"
                dist_table.append([
                    interval,
                    str(b["count"]),
                    f"{b['pct']:.1f}%",
                ])
            dist_tables.append(
                f"── {label} 分布 ({name},{len(values)}条) ──\n"
                + _format_table(dist_headers, dist_table)
            )

        print(summary_str)
        print()
        if dist_tables:
            print(_join_tables_side_by_side(dist_tables))
        print()


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
    p_show.add_argument("--all", action="store_true",
                        help="Show all 6 spread types (default: only 高-开 and 开-低)")

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
