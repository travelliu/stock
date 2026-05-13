import pytest
import pandas as pd
from unittest.mock import patch, MagicMock

from fetcher import compute_spreads, fetch_daily


class TestComputeSpreads:
    def test_basic_spread_computation(self):
        df = pd.DataFrame([
            {"open": 10.0, "high": 11.0, "low": 9.5, "close": 10.5},
        ])
        result = compute_spreads(df)
        row = result.iloc[0]
        assert row["spread_oh"] == pytest.approx(1.0)    # high - open
        assert row["spread_ol"] == pytest.approx(0.5)    # open - low
        assert row["spread_hl"] == pytest.approx(1.5)    # high - low
        assert row["spread_oc"] == pytest.approx(0.5)    # |open - close|
        assert row["spread_hc"] == pytest.approx(0.5)    # |high - close|
        assert row["spread_lc"] == pytest.approx(1.0)    # |low - close|

    def test_preserves_original_columns(self):
        df = pd.DataFrame([
            {"ts_code": "603778.SH", "trade_date": "2025-05-12",
             "open": 10.0, "high": 11.0, "low": 9.5, "close": 10.5,
             "vol": 1000.0, "amount": 10000.0},
        ])
        result = compute_spreads(df)
        assert "ts_code" in result.columns
        assert "trade_date" in result.columns
        assert result.iloc[0]["ts_code"] == "603778.SH"

    def test_multiple_rows(self):
        df = pd.DataFrame([
            {"open": 10.0, "high": 11.0, "low": 9.5, "close": 10.5},
            {"open": 20.0, "high": 21.0, "low": 19.0, "close": 20.5},
        ])
        result = compute_spreads(df)
        assert len(result) == 2
        assert result.iloc[1]["spread_oh"] == pytest.approx(1.0)


class TestFetchDaily:
    @patch("fetcher.tushare")
    def test_fetch_calls_api_with_correct_params(self, mock_ts):
        mock_pro = MagicMock()
        mock_ts.pro_api.return_value = mock_pro
        mock_pro.daily.return_value = pd.DataFrame([
            {"ts_code": "603778.SH", "trade_date": "20250512",
             "open": 10.0, "high": 11.0, "low": 9.5, "close": 10.5,
             "vol": 1000.0, "amount": 10000.0},
        ])
        result = fetch_daily("603778", days=30)
        mock_pro.daily.assert_called_once()
        call_kwargs = mock_pro.daily.call_args[1]
        assert call_kwargs["ts_code"] == "603778.SH"

    @patch("fetcher.tushare")
    def test_fetch_converts_date_format(self, mock_ts):
        mock_pro = MagicMock()
        mock_ts.pro_api.return_value = mock_pro
        mock_pro.daily.return_value = pd.DataFrame([
            {"ts_code": "603778.SH", "trade_date": "20250512",
             "open": 10.0, "high": 11.0, "low": 9.5, "close": 10.5,
             "vol": 1000.0, "amount": 10000.0},
        ])
        result = fetch_daily("603778", days=30)
        assert result.iloc[0]["trade_date"] == "2025-05-12"

    @patch("fetcher.tushare")
    def test_fetch_with_start_date(self, mock_ts):
        mock_pro = MagicMock()
        mock_ts.pro_api.return_value = mock_pro
        mock_pro.daily.return_value = pd.DataFrame([
            {"ts_code": "603778.SH", "trade_date": "20250513",
             "open": 10.0, "high": 11.0, "low": 9.5, "close": 10.5,
             "vol": 1000.0, "amount": 10000.0},
        ])
        result = fetch_daily("603778", start_date="2025-05-12")
        call_kwargs = mock_pro.daily.call_args[1]
        assert call_kwargs["start_date"] == "20250512"
