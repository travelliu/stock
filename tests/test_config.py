# tests/test_config.py
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from config import (
    PROJECT_ROOT,
    DATA_DIR,
    DAILY_DIR,
    AUCTION_DIR,
    STOCKS_LIST_PATH,
    VOLUME_RATIO_THRESHOLD,
    OPEN_CHANGE_THRESHOLD,
    TOP_VOLUME_COUNT,
    AUCTION_HISTORY_DAYS,
    RETRY_MAX_ATTEMPTS,
    RETRY_BACKOFF_BASE,
    CONCURRENT_WORKERS,
)


def test_project_root_exists():
    assert PROJECT_ROOT.exists()


def test_data_dirs_are_under_project():
    assert str(DATA_DIR).startswith(str(PROJECT_ROOT))
    assert str(DAILY_DIR).startswith(str(DATA_DIR))
    assert str(AUCTION_DIR).startswith(str(DATA_DIR))


def test_stocks_list_path_is_csv():
    assert STOCKS_LIST_PATH.suffix == ".csv"
    assert STOCKS_LIST_PATH.name == "stocks_list.csv"


def test_threshold_values_are_reasonable():
    assert VOLUME_RATIO_THRESHOLD > 0
    assert 0 < OPEN_CHANGE_THRESHOLD < 20
    assert TOP_VOLUME_COUNT > 0
    assert AUCTION_HISTORY_DAYS == 5
    assert RETRY_MAX_ATTEMPTS >= 1
    assert RETRY_BACKOFF_BASE >= 1
    assert CONCURRENT_WORKERS >= 1
