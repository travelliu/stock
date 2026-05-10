# init_data.py
"""One-time initialization: fetch stock list and historical daily data."""

import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime, timedelta

import akshare as ak
import pandas as pd

from config import (
    CONCURRENT_WORKERS,
    DAILY_DIR,
    DAILY_HISTORY_DAYS,
    RETRY_BACKOFF_BASE,
    RETRY_MAX_ATTEMPTS,
    STOCKS_LIST_PATH,
)


def retry_call(fn, *args, **kwargs):
    """Call fn with retry and exponential backoff."""
    for attempt in range(1, RETRY_MAX_ATTEMPTS + 1):
        try:
            return fn(*args, **kwargs)
        except Exception as e:
            if attempt == RETRY_MAX_ATTEMPTS:
                print(f"  [FAIL] {fn.__name__} after {attempt} attempts: {e}")
                return None
            wait = RETRY_BACKOFF_BASE ** attempt
            print(f"  [RETRY] {fn.__name__} attempt {attempt}, waiting {wait}s: {e}")
            time.sleep(wait)


def fetch_stock_list():
    """Fetch all A-share stock codes and names, save to CSV."""
    print("Fetching stock list...")
    df = retry_call(ak.stock_zh_a_spot_em)
    if df is None:
        print("Failed to fetch stock list.")
        return None

    stock_list = df[["代码", "名称"]].copy()
    stock_list.to_csv(STOCKS_LIST_PATH, index=False, encoding="utf-8-sig")
    print(f"Saved {len(stock_list)} stocks to {STOCKS_LIST_PATH}")
    return stock_list


def fetch_one_daily(symbol: str, name: str):
    """Fetch daily history for one stock and save to CSV."""
    start_date = (datetime.now() - timedelta(days=DAILY_HISTORY_DAYS + 10)).strftime("%Y%m%d")
    end_date = datetime.now().strftime("%Y%m%d")

    df = retry_call(
        ak.stock_zh_a_hist,
        symbol=symbol,
        period="daily",
        start_date=start_date,
        end_date=end_date,
        adjust="qfq",
    )
    if df is None or df.empty:
        print(f"  [SKIP] {symbol} {name}: no data")
        return False

    df.to_csv(DAILY_DIR / f"{symbol}.csv", index=False, encoding="utf-8-sig")
    return True


def init_all_daily(stock_list: pd.DataFrame):
    """Fetch daily data for all stocks concurrently."""
    total = len(stock_list)
    success = 0
    print(f"Fetching daily data for {total} stocks...")

    with ThreadPoolExecutor(max_workers=CONCURRENT_WORKERS) as executor:
        futures = {
            executor.submit(fetch_one_daily, row["代码"], row["名称"]): idx
            for idx, row in stock_list.iterrows()
        }
        for future in as_completed(futures):
            idx = futures[future]
            try:
                if future.result():
                    success += 1
            except Exception as e:
                print(f"  [ERROR] stock index {idx}: {e}")

            done_count = success + (total - len([f for f in futures if not f.done()]))
            if done_count % 100 == 0:
                print(f"  Progress: {done_count}/{total}")

    print(f"Completed: {success}/{total} stocks fetched successfully")


def main():
    stock_list = fetch_stock_list()
    if stock_list is not None:
        init_all_daily(stock_list)
    print("Initialization complete.")


if __name__ == "__main__":
    main()
