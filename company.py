"""Stock company name resolution with JSON file caching."""

import json
from pathlib import Path

from config import DATA_DIR, TUSHARE_TOKEN, to_tushare_code

NAMES_CACHE = DATA_DIR / "stock_names.json"


def get_stock_name(stock_code: str) -> str:
    """Return stock name, auto-fetching from Tushare on cache miss.

    Args:
        stock_code: plain 6-digit code like '603778'

    Returns:
        Stock short name (e.g., '千金药业'), or empty string on failure.
    """
    names = _load_cache()
    if stock_code in names:
        return names[stock_code]

    name = _fetch_name(stock_code)
    if name:
        names[stock_code] = name
        _save_cache(names)
    return name


def _load_cache() -> dict[str, str]:
    """Load stock name cache from JSON file."""
    if not NAMES_CACHE.exists():
        return {}
    try:
        with open(NAMES_CACHE, "r", encoding="utf-8") as f:
            return json.load(f)
    except (json.JSONDecodeError, OSError):
        return {}


def _save_cache(names: dict[str, str]) -> None:
    """Write stock name cache to JSON file."""
    with open(NAMES_CACHE, "w", encoding="utf-8") as f:
        json.dump(names, f, ensure_ascii=False, indent=2)


def _fetch_name(stock_code: str) -> str:
    """Fetch stock short name from Tushare stock_company API."""
    import tushare

    ts_code = to_tushare_code(stock_code)
    pro = tushare.pro_api(TUSHARE_TOKEN)
    df = pro.stock_company(ts_code=ts_code, fields="ts_code,short_name")
    if df is not None and not df.empty:
        return str(df.iloc[0]["short_name"])
    return ""
