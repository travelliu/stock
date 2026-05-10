# config.py
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parent
DATA_DIR = PROJECT_ROOT / "data"
DAILY_DIR = DATA_DIR / "daily"
AUCTION_DIR = DATA_DIR / "auction"
STOCKS_LIST_PATH = DATA_DIR / "stocks_list.csv"

# Screening thresholds
VOLUME_RATIO_THRESHOLD = 2.0
OPEN_CHANGE_THRESHOLD = 3.0
TOP_VOLUME_COUNT = 50

# Analysis parameters
AUCTION_HISTORY_DAYS = 5
DAILY_HISTORY_DAYS = 30

# API retry config
RETRY_MAX_ATTEMPTS = 3
RETRY_BACKOFF_BASE = 2

# Concurrency
CONCURRENT_WORKERS = 5

# Ensure data dirs exist
for _d in (DATA_DIR, DAILY_DIR, AUCTION_DIR):
    _d.mkdir(parents=True, exist_ok=True)
