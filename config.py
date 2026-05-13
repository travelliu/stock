import os

from dotenv import load_dotenv
from pathlib import Path

# Directories
PROJECT_DIR = Path(__file__).resolve().parent
DATA_DIR = PROJECT_DIR / "data"
DATA_DIR.mkdir(exist_ok=True)

# Load .env
load_dotenv(PROJECT_DIR / ".env")

# tushare
TUSHARE_TOKEN = os.environ.get("TUSHARE_TOKEN", "")

# Database
DB_PATH = DATA_DIR / "stock_daily.db"

# Default stocks (comma-separated in env var DEFAULT_STOCKS, or hardcoded fallback)
DEFAULT_STOCKS = [
    s.strip() for s in os.environ.get(
        "DEFAULT_STOCKS",
        "600537,603778,000890,600186,300342,000593,600821,300476",
    ).split(",") if s.strip()
]

# Fetch defaults
DEFAULT_FETCH_DAYS = 180

# Market determination by 3-digit prefix
_SH_PREFIXES = {"600", "601", "603", "605", "688", "900",
                "510", "511", "512", "513", "515"}
_SZ_PREFIXES = {"000", "001", "002", "300", "200", "159"}


def to_tushare_code(code: str) -> str:
    """Convert plain stock code to tushare format.

    Rules:
        SH (上交所): 600/601/603/605 (主板), 688 (科创板),
                     900 (B股), 510/511/512/513/515 (ETF)
        SZ (深交所): 000/001 (主板), 002 (中小板), 300 (创业板),
                     200 (B股), 159 (ETF)
        IPO subscription codes (730, 732, etc.) raise ValueError.

    Examples:
        '600537' -> '600537.SH'
        '000890' -> '000890.SZ'
        '688001' -> '688001.SH'
        '159915' -> '159915.SZ'

    Already-suffixed codes pass through unchanged.
    """
    if "." in code:
        return code
    if len(code) < 6:
        raise ValueError(
            f"Invalid stock code '{code}': must be a 6-digit code"
        )
    prefix = code[:3]
    if prefix in _SH_PREFIXES:
        return f"{code}.SH"
    if prefix in _SZ_PREFIXES:
        return f"{code}.SZ"
    raise ValueError(
        f"Cannot determine market for stock code '{code}' "
        f"(prefix '{prefix}' is not a known SH/SZ prefix)"
    )
