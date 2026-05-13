#!/usr/bin/env python3
import argparse
from datetime import datetime, timedelta

from analysis import StockAnalyzer
from company import get_stock_name
from config import DEFAULT_FETCH_DAYS, DEFAULT_STOCKS, DB_PATH
from db import DailyDB
from fetcher import fetch_daily


def cmd_fetch(args: argparse.Namespace) -> None:
    stocks_input = getattr(args, "stocks", None)
    if stocks_input:
        stocks = [s.strip() for s in stocks_input.split(",") if s.strip()]
    else:
        stocks = DEFAULT_STOCKS
        print(f"Using DEFAULT_STOCKS: {', '.join(stocks)}")

    db = DailyDB(DB_PATH)
    db.init()

    for code in stocks:
        name = get_stock_name(code)
        label = f"{code} {name}" if name else code
        print(f"Fetching {label} ...")

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
                # No existing data, fetch full history
                days = 3650
                start_date = None
                print(f"  No existing data, fetching full history ...")

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

    all_rows = db.query_daily(stock, "2000-01-01", end_date)
    analyzer = StockAnalyzer(
        stock,
        all_rows=all_rows or [],
        start_date=start_date,
        end_date=end_date,
        show_all=show_all,
        open_price=args.open,
        actual_low=args.low,
        actual_high=args.high,
        actual_close=args.close,
    )
    analyzer.show()


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Stock Daily Data CLI (Tushare)"
    )
    sub = parser.add_subparsers(dest="command")

    # fetch
    p_fetch = sub.add_parser("fetch", help="Fetch daily data from tushare")
    p_fetch.add_argument("--stocks", type=str, default=None,
                         help="Comma-separated stock codes (e.g., 603778,000890). "
                              "Default: use DEFAULT_STOCKS from env")
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
    p_show.add_argument("--open", type=float, default=None,
                        help="Today's opening price. When provided, outputs a trading plan report.")
    p_show.add_argument("--low", type=float, default=None,
                        help="Today's lowest price. Used for reverse calculation in reference table.")
    p_show.add_argument("--high", type=float, default=None,
                        help="Today's highest price. Used for reverse calculation in reference table.")
    p_show.add_argument("--close", type=float, default=None,
                        help="Today's closing price. When provided, shown in header instead of predicted value.")

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
