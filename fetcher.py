from datetime import datetime, timedelta

import pandas as pd
import tushare

from config import TUSHARE_TOKEN, to_tushare_code


def compute_spreads(df: pd.DataFrame) -> pd.DataFrame:
    """Compute 6 price-spread columns and append to DataFrame.

    All spreads are stored as absolute values.

    spread_oh = |high - open|
    spread_ol = |open - low|
    spread_hl = |high - low|
    spread_oc = |open - close|
    spread_hc = |high - close|
    spread_lc = |low - close|
    """
    result = df.copy()
    result["spread_oh"] = (result["high"] - result["open"]).abs()
    result["spread_ol"] = (result["open"] - result["low"]).abs()
    result["spread_hl"] = (result["high"] - result["low"]).abs()
    result["spread_oc"] = (result["open"] - result["close"]).abs()
    result["spread_hc"] = (result["high"] - result["close"]).abs()
    result["spread_lc"] = (result["low"] - result["close"]).abs()
    return result


def fetch_daily(
    stock_code: str,
    start_date: str | None = None,
    days: int | None = None,
) -> pd.DataFrame:
    """Fetch daily OHLCV from tushare and compute spreads.

    Args:
        stock_code: plain code like '603778'
        start_date: fetch data from this date (YYYY-MM-DD), exclusive.
                    Used for incremental fetch — data on or after this date.
        days: lookback days from today. Used only when start_date is None.
              Defaults to DEFAULT_FETCH_DAYS (180).

    Returns:
        DataFrame with OHLCV + spread columns, trade_date in 'YYYY-MM-DD' format.
    """
    from config import DEFAULT_FETCH_DAYS

    ts_code = to_tushare_code(stock_code)
    end_date = datetime.now().strftime("%Y%m%d")

    if start_date:
        # Incremental: start from the day after the existing max date
        sd = datetime.strptime(start_date, "%Y-%m-%d")
        api_start = sd.strftime("%Y%m%d")
    else:
        lookback = days if days is not None else DEFAULT_FETCH_DAYS
        api_start = (datetime.now() - timedelta(days=lookback)).strftime("%Y%m%d")

    pro = tushare.pro_api(TUSHARE_TOKEN)
    df = pro.daily(ts_code=ts_code, start_date=api_start, end_date=end_date)

    if df is None or df.empty:
        return pd.DataFrame()

    # Convert tushare date format 'YYYYMMDD' -> 'YYYY-MM-DD'
    df["trade_date"] = pd.to_datetime(df["trade_date"], format="%Y%m%d").dt.strftime(
        "%Y-%m-%d"
    )

    # Sort by date ascending
    df = df.sort_values("trade_date").reset_index(drop=True)

    return compute_spreads(df)
