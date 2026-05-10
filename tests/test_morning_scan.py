# tests/test_morning_scan.py
import sys
from pathlib import Path
from unittest.mock import patch

import pandas as pd
import pytest

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from morning_scan import screen_candidates


def test_screen_candidates_basic():
    spot_df = pd.DataFrame([
        {"代码": "000001", "名称": "平安银行", "量比": 3.5, "涨跌幅": 1.5, "成交额": 5e8},
        {"代码": "000002", "名称": "万科A", "量比": 1.0, "涨跌幅": 0.5, "成交额": 2e8},
        {"代码": "000003", "名称": "异动股", "量比": 0.8, "涨跌幅": 5.0, "成交额": 1e8},
        {"代码": "000004", "名称": "巨量股", "量比": 0.5, "涨跌幅": -0.5, "成交额": 1.5e9},
        {"代码": "000005", "名称": "普通股", "量比": 0.5, "涨跌幅": 0.2, "成交额": 5e7},
        {"代码": "000006", "名称": "平淡股", "量比": 0.3, "涨跌幅": -1.0, "成交额": 3e7},
    ])

    # Patch the constants in the morning_scan module namespace
    with patch("morning_scan.TOP_VOLUME_COUNT", 2):
        candidates = screen_candidates(spot_df)

    codes = candidates["代码"].tolist()
    assert "000001" in codes  # volume ratio > 2
    assert "000003" in codes  # open change > 3%
    # 000004 (1.5e9) and 000002 (2e8) are top 2 by volume
    assert "000005" not in codes  # nothing special
    assert "000006" not in codes  # nothing special


def test_screen_candidates_empty():
    spot_df = pd.DataFrame(columns=["代码", "名称", "量比", "涨跌幅", "成交额"])
    candidates = screen_candidates(spot_df)
    assert candidates.empty
