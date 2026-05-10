# morning_scan.py
"""Morning auction scan: screen candidates from spot data, fetch auction details."""

import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import date

import akshare as ak
import pandas as pd

from config import (
    AUCTION_DIR,
    CONCURRENT_WORKERS,
    OPEN_CHANGE_THRESHOLD,
    RETRY_BACKOFF_BASE,
    RETRY_MAX_ATTEMPTS,
    TOP_VOLUME_COUNT,
    VOLUME_RATIO_THRESHOLD,
)
from analyzer import build_report, format_report, load_auction_history


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


def fetch_spot_data():
    """Fetch real-time spot data for all A-shares."""
    print("Fetching real-time spot data...")
    df = retry_call(ak.stock_zh_a_spot_em)
    if df is None:
        print("Failed to fetch spot data.")
        return pd.DataFrame()
    print(f"Fetched {len(df)} stocks.")
    return df


def screen_candidates(spot_df: pd.DataFrame) -> pd.DataFrame:
    """Screen stocks matching at least one of the filter criteria."""
    if spot_df.empty:
        return pd.DataFrame()

    mask_volume_ratio = spot_df["量比"] > VOLUME_RATIO_THRESHOLD
    mask_open_change = spot_df["涨跌幅"] > OPEN_CHANGE_THRESHOLD

    top_n = min(TOP_VOLUME_COUNT, len(spot_df))
    top_volume_threshold = spot_df["成交额"].nlargest(top_n).min()
    mask_top_volume = spot_df["成交额"] >= top_volume_threshold

    combined = mask_volume_ratio | mask_open_change | mask_top_volume
    return spot_df[combined].copy()


def fetch_auction_data(symbol: str):
    """Fetch call auction detail for one stock."""
    df = retry_call(ak.stock_zh_a_call_auction_em, symbol=symbol)
    if df is None or df.empty:
        return None
    return df


def build_auction_record(symbol: str, name: str, auction_df: pd.DataFrame) -> dict:
    """Extract key fields from raw auction data into a flat record."""
    if auction_df is None or auction_df.empty:
        return None

    last = auction_df.iloc[-1]

    return {
        "股票代码": symbol,
        "股票名称": name,
        "竞价成交价": last.get("成交价", float("nan")),
        "竞价成交量": last.get("成交量", 0),
        "竞价成交额": last.get("成交额", 0),
        "未匹配量": last.get("未匹配量", 0),
        "粗筛标记": True,
    }


def save_auction_data(records: list, today: date):
    """Save today's auction records to CSV."""
    if not records:
        print("No auction records to save.")
        return

    df = pd.DataFrame(records)
    csv_path = AUCTION_DIR / f"{today}.csv"
    df.to_csv(csv_path, index=False, encoding="utf-8-sig")
    print(f"Saved {len(records)} auction records to {csv_path}")


def run_morning_scan():
    """Main morning scan pipeline."""
    today = date.today()

    # Step 1: Fetch spot data
    spot_df = fetch_spot_data()
    if spot_df.empty:
        return

    # Step 2: Screen candidates
    candidates = screen_candidates(spot_df)
    print(f"Screened {len(candidates)} candidates from {len(spot_df)} stocks.")

    if candidates.empty:
        print("No candidates found.")
        return

    # Step 3: Fetch auction data for candidates
    records = []
    with ThreadPoolExecutor(max_workers=CONCURRENT_WORKERS) as executor:
        futures = {
            executor.submit(fetch_auction_data, row["代码"]): (row["代码"], row["名称"])
            for _, row in candidates.iterrows()
        }
        for future in as_completed(futures):
            symbol, name = futures[future]
            try:
                auction_df = future.result()
                record = build_auction_record(symbol, name, auction_df)
                if record:
                    records.append(record)
            except Exception as e:
                print(f"  [ERROR] {symbol} {name}: {e}")

    # Step 4: Save auction data
    save_auction_data(records, today)

    # Step 5: Analyze and output
    if not records:
        print("No valid auction data collected.")
        return

    today_auction = pd.DataFrame(records)
    auction_history = load_auction_history(AUCTION_DIR, today)
    report = build_report(
        today_auction=today_auction,
        auction_history=auction_history,
        daily_dir=AUCTION_DIR.parent / "daily",
        today=today,
    )
    print("\n=== 集合竞价量比分析报告 ===")
    print(f"日期: {today} | 候选股数: {len(report)}")
    format_report(report)


if __name__ == "__main__":
    run_morning_scan()
