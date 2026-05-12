"""Tests for company module — stock name resolution with JSON caching."""

import json
import sys
from pathlib import Path
from unittest.mock import patch, MagicMock

import pandas as pd
import pytest

from company import get_stock_name, _load_cache, _save_cache, _fetch_name, NAMES_CACHE


def _mock_tushare_module():
    """Create a mock tushare module for patching into sys.modules."""
    return MagicMock()


class TestLoadCache:
    def test_returns_empty_when_file_missing(self, tmp_path: Path, monkeypatch):
        monkeypatch.setattr("company.NAMES_CACHE", tmp_path / "missing.json")
        result = _load_cache()
        assert result == {}

    def test_returns_empty_when_file_corrupt(self, tmp_path: Path, monkeypatch):
        cache_file = tmp_path / "bad.json"
        cache_file.write_text("not valid json{{{", encoding="utf-8")
        monkeypatch.setattr("company.NAMES_CACHE", cache_file)
        result = _load_cache()
        assert result == {}

    def test_loads_valid_cache(self, tmp_path: Path, monkeypatch):
        cache_file = tmp_path / "stock_names.json"
        cache_file.write_text(
            json.dumps({"603778": "千金药业"}, ensure_ascii=False),
            encoding="utf-8",
        )
        monkeypatch.setattr("company.NAMES_CACHE", cache_file)
        result = _load_cache()
        assert result == {"603778": "千金药业"}


class TestSaveCache:
    def test_writes_json_file(self, tmp_path: Path, monkeypatch):
        cache_file = tmp_path / "stock_names.json"
        monkeypatch.setattr("company.NAMES_CACHE", cache_file)
        _save_cache({"000890": "法尔胜", "603778": "千金药业"})
        data = json.loads(cache_file.read_text(encoding="utf-8"))
        assert data == {"000890": "法尔胜", "603778": "千金药业"}

    def test_round_trip(self, tmp_path: Path, monkeypatch):
        cache_file = tmp_path / "stock_names.json"
        monkeypatch.setattr("company.NAMES_CACHE", cache_file)
        original = {"603778": "千金药业", "000890": "法尔胜"}
        _save_cache(original)
        loaded = _load_cache()
        assert loaded == original


class TestFetchName:
    @patch.dict(sys.modules, {"tushare": _mock_tushare_module()})
    def test_returns_short_name(self):
        mock_ts = sys.modules["tushare"]
        mock_pro = MagicMock()
        mock_ts.pro_api.return_value = mock_pro
        mock_pro.stock_company.return_value = pd.DataFrame([
            {"ts_code": "603778.SH", "short_name": "千金药业"},
        ])
        result = _fetch_name("603778")
        assert result == "千金药业"
        mock_pro.stock_company.assert_called_once_with(
            ts_code="603778.SH", fields="ts_code,short_name"
        )

    @patch.dict(sys.modules, {"tushare": _mock_tushare_module()})
    def test_returns_empty_on_empty_response(self):
        mock_ts = sys.modules["tushare"]
        mock_pro = MagicMock()
        mock_ts.pro_api.return_value = mock_pro
        mock_pro.stock_company.return_value = pd.DataFrame()
        result = _fetch_name("603778")
        assert result == ""

    @patch.dict(sys.modules, {"tushare": _mock_tushare_module()})
    def test_returns_empty_on_none_response(self):
        mock_ts = sys.modules["tushare"]
        mock_pro = MagicMock()
        mock_ts.pro_api.return_value = mock_pro
        mock_pro.stock_company.return_value = None
        result = _fetch_name("603778")
        assert result == ""


class TestGetStockName:
    def test_returns_from_cache(self, tmp_path: Path, monkeypatch):
        cache_file = tmp_path / "stock_names.json"
        cache_file.write_text(
            json.dumps({"603778": "千金药业"}, ensure_ascii=False),
            encoding="utf-8",
        )
        monkeypatch.setattr("company.NAMES_CACHE", cache_file)
        # No API mock needed — should hit cache
        result = get_stock_name("603778")
        assert result == "千金药业"

    @patch("company._fetch_name", return_value="法尔胜")
    def test_fetches_on_cache_miss_and_saves(self, mock_fetch, tmp_path: Path, monkeypatch):
        cache_file = tmp_path / "stock_names.json"
        monkeypatch.setattr("company.NAMES_CACHE", cache_file)
        result = get_stock_name("000890")
        assert result == "法尔胜"
        mock_fetch.assert_called_once_with("000890")
        # Verify cache was written
        saved = json.loads(cache_file.read_text(encoding="utf-8"))
        assert saved == {"000890": "法尔胜"}

    @patch("company._fetch_name", return_value="")
    def test_returns_empty_and_does_not_cache_failure(self, mock_fetch, tmp_path: Path, monkeypatch):
        cache_file = tmp_path / "stock_names.json"
        monkeypatch.setattr("company.NAMES_CACHE", cache_file)
        result = get_stock_name("999999")
        assert result == ""
        # Cache file should not exist (no write on failure)
        assert not cache_file.exists()
