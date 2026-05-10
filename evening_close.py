# evening_close.py
"""Evening update: fetch today's daily OHLCV data and append to local CSVs."""

import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime

import akshare as ak
import pandas as pd

from config import (
    CONCURRENT_WORKERS,
    DAILY_DIR,
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


def update_one_stock(symbol: str, name: str):
    """Fetch today's daily data for one stock and append to existing CSV."""
    today = datetime.now().strftime("%Y%m%d")

    df = retry_call(
        ak.stock_zh_a_hist,
        symbol=symbol,
        period="daily",
        start_date=today,
        end_date=today,
        adjust="qfq",
    )
    if df is None or df.empty:
        return False

    csv_path = DAILY_DIR / f"{symbol}.csv"
    if csv_path.exists():
        existing = pd.read_csv(csv_path)
        new_date = df.iloc[0]["日期"]
        if new_date in existing["日期"].values:
            return True
        combined = pd.concat([existing, df], ignore_index=True)
        combined.to_csv(csv_path, index=False, encoding="utf-8-sig")
    else:
        df.to_csv(csv_path, index=False, encoding="utf-8-sig")
    return True


def run_evening_update():
    """Main evening update pipeline."""
    if not STOCKS_LIST_PATH.exists():
        print(f"Stock list not found at {STOCKS_LIST_PATH}. Run init_data.py first.")
        return

    stock_list = pd.read_csv(STOCKS_LIST_PATH, dtype={"代码": str})
    total = len(stock_list)
    success = 0

    print(f"Updating daily data for {total} stocks...")

    with ThreadPoolExecutor(max_workers=CONCURRENT_WORKERS) as executor:
        futures = {
            executor.submit(update_one_stock, row["代码"], row["名称"]): idx
            for idx, row in stock_list.iterrows()
        }
        for future in as_completed(futures):
            idx = futures[future]
            try:
                if future.result():
                    success += 1
            except Exception as e:
                print(f"  [ERROR] index {idx}: {e}")

    print(f"Evening update complete: {success}/{total} stocks updated.")


if __name__ == "__main__":
    run_evening_update()
