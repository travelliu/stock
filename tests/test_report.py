import statistics
import pytest
from report import (
    _compute_window_means,
    _compute_composite_means,
    MODEL_SPREAD_KEYS,
    _display_width,
    _rpad,
    _format_table,
    _format_header,
    _build_spread_model_table,
    MODEL_SPREAD_LABELS,
    _build_reference_table,
    build_trading_plan,
)


class TestComputeWindowMeans:
    def test_basic(self):
        rows = [
            {"trade_date": "2024-01-04", "spread_oh": 1.0, "spread_ol": 0.5,
             "spread_hl": 1.5, "spread_oc": 0.5, "spread_hc": 0.5, "spread_lc": 1.0},
            {"trade_date": "2024-01-03", "spread_oh": 2.0, "spread_ol": 1.0,
             "spread_hl": 3.0, "spread_oc": 1.0, "spread_hc": 1.0, "spread_lc": 2.0},
            {"trade_date": "2024-01-02", "spread_oh": 3.0, "spread_ol": 1.5,
             "spread_hl": 4.5, "spread_oc": 1.5, "spread_hc": 1.5, "spread_lc": 3.0},
        ]
        means = _compute_window_means(rows)
        assert means["历史"]["spread_oh"] == pytest.approx(2.0)
        assert means["历史"]["spread_ol"] == pytest.approx(1.0)
        assert means["近3月"]["spread_oh"] == pytest.approx(2.0)
        assert means["近2周"]["spread_ol"] == pytest.approx(1.0)

    def test_empty_rows(self):
        means = _compute_window_means([])
        for wname in ["历史", "近3月", "近1月", "近2周"]:
            assert all(v is None for v in means[wname].values())

    def test_rows_with_none_values(self):
        rows = [
            {"trade_date": "2024-01-03", "spread_oh": 1.0, "spread_ol": None,
             "spread_hl": 1.5, "spread_oc": None, "spread_hc": 0.5, "spread_lc": 1.0},
            {"trade_date": "2024-01-02", "spread_oh": 2.0, "spread_ol": 1.0,
             "spread_hl": None, "spread_oc": 1.0, "spread_hc": None, "spread_lc": 2.0},
        ]
        means = _compute_window_means(rows)
        assert means["历史"]["spread_oh"] == pytest.approx(1.5)
        assert means["历史"]["spread_ol"] == pytest.approx(1.0)
        assert means["历史"]["spread_hl"] == pytest.approx(1.5)
        assert means["历史"]["spread_oc"] == pytest.approx(1.0)
        assert means["历史"]["spread_hc"] == pytest.approx(0.5)
        assert means["历史"]["spread_lc"] == pytest.approx(1.5)
        # All rows fall into every window (only 2 rows), so all windows should match
        assert means["近3月"]["spread_oh"] == pytest.approx(1.5)
        assert means["近1月"]["spread_oh"] == pytest.approx(1.5)
        assert means["近2周"]["spread_oh"] == pytest.approx(1.5)


class TestComputeCompositeMeans:
    def test_basic(self):
        window_means = {
            "历史": {"spread_oh": 4.0, "spread_ol": 3.0, "spread_hl": 8.0,
                     "spread_oc": 4.0, "spread_hc": 4.0, "spread_lc": 3.0},
            "近3月": {"spread_oh": 2.0, "spread_ol": 1.0, "spread_hl": 4.0,
                      "spread_oc": 2.0, "spread_hc": 2.0, "spread_lc": 1.0},
            "近1月": {"spread_oh": 1.0, "spread_ol": 0.5, "spread_hl": 2.0,
                      "spread_oc": 1.0, "spread_hc": 1.0, "spread_lc": 0.5},
            "近2周": {"spread_oh": 0.5, "spread_ol": 0.25, "spread_hl": 1.0,
                      "spread_oc": 0.5, "spread_hc": 0.5, "spread_lc": 0.25},
        }
        composite = _compute_composite_means(window_means)
        assert composite["spread_oh"] == pytest.approx(1.875)
        assert composite["spread_ol"] == pytest.approx(1.1875)

    def test_with_none_values(self):
        window_means = {
            "历史": {"spread_oh": 4.0, "spread_ol": None, "spread_hl": 8.0,
                     "spread_oc": 4.0, "spread_hc": 4.0, "spread_lc": None},
            "近3月": {"spread_oh": 2.0, "spread_ol": None, "spread_hl": 4.0,
                      "spread_oc": 2.0, "spread_hc": 2.0, "spread_lc": None},
            "近1月": {"spread_oh": 1.0, "spread_ol": None, "spread_hl": 2.0,
                      "spread_oc": 1.0, "spread_hc": 1.0, "spread_lc": None},
            "近2周": {"spread_oh": 0.5, "spread_ol": None, "spread_hl": 1.0,
                      "spread_oc": 0.5, "spread_hc": 0.5, "spread_lc": None},
        }
        composite = _compute_composite_means(window_means)
        assert composite["spread_oh"] == pytest.approx(1.875)
        assert composite["spread_ol"] == 0.0

    def test_all_none_values(self):
        window_means = {
            "历史": {"spread_oh": None, "spread_ol": None, "spread_hl": None,
                     "spread_oc": None, "spread_hc": None, "spread_lc": None},
            "近3月": {"spread_oh": None, "spread_ol": None, "spread_hl": None,
                      "spread_oc": None, "spread_hc": None, "spread_lc": None},
            "近1月": {"spread_oh": None, "spread_ol": None, "spread_hl": None,
                      "spread_oc": None, "spread_hc": None, "spread_lc": None},
            "近2周": {"spread_oh": None, "spread_ol": None, "spread_hl": None,
                      "spread_oc": None, "spread_hc": None, "spread_lc": None},
        }
        composite = _compute_composite_means(window_means)
        assert composite["spread_oh"] == 0.0
        assert composite["spread_ol"] == 0.0


class TestTableFormatting:
    def test_display_width_english(self):
        assert _display_width("hello") == 5

    def test_display_width_chinese(self):
        assert _display_width("开盘") == 4

    def test_display_width_mixed(self):
        assert _display_width("开A") == 3

    def test_format_table_basic(self):
        headers = ["时段", "数值"]
        rows = [["历史", "1.23"], ["近1月", "0.45"]]
        table = _format_table(headers, rows)
        assert "时段" in table
        assert "历史" in table
        assert "+" in table
        lines = table.split("\n")
        assert len(lines) == 6


class TestFormatHeader:
    def test_basic(self):
        composite = {"spread_oh": 3.77, "spread_ol": 3.82,
                     "spread_oc": 4.07, "spread_hl": 8.21,
                     "spread_hc": 4.07, "spread_lc": 3.25}
        header = _format_header(200.0, composite)
        assert "开盘价: 200.00" in header
        assert "最高价: 203.77" in header
        assert "最低价: 196.18" in header
        assert "收盘价: 195.93" in header


class TestBuildSpreadModelTable:
    def test_structure(self):
        window_means = {
            "历史": {k: 1.0 for k in MODEL_SPREAD_KEYS},
            "近3月": {k: 0.5 for k in MODEL_SPREAD_KEYS},
            "近1月": {k: 0.3 for k in MODEL_SPREAD_KEYS},
            "近2周": {k: 0.2 for k in MODEL_SPREAD_KEYS},
        }
        composite = {k: 0.5 for k in MODEL_SPREAD_KEYS}
        table = _build_spread_model_table(window_means, composite)
        assert "时段" in table
        assert "开盘与最高价" in table
        assert "综合均值" in table
        assert "1.00" in table
        lines = table.split("\n")
        assert len(lines) == 9

    def test_with_none(self):
        window_means = {
            "历史": {k: 1.0 for k in MODEL_SPREAD_KEYS},
            "近3月": {k: None for k in MODEL_SPREAD_KEYS},
            "近1月": {k: None for k in MODEL_SPREAD_KEYS},
            "近2周": {k: None for k in MODEL_SPREAD_KEYS},
        }
        composite = {k: 1.0 for k in MODEL_SPREAD_KEYS}
        table = _build_spread_model_table(window_means, composite)
        assert "-" in table


class TestBuildReferenceTable:
    def test_high_low_close_rows(self):
        window_means = {
            "历史": {"spread_oh": 3.77, "spread_ol": 3.82, "spread_hc": 4.07,
                     "spread_lc": 3.25, "spread_hl": 8.21, "spread_oc": 4.07},
            "近3月": {"spread_oh": 2.41, "spread_ol": 2.50, "spread_hc": 2.69,
                      "spread_lc": 1.72, "spread_hl": 4.91, "spread_oc": 2.25},
            "近1月": {"spread_oh": 1.72, "spread_ol": 1.59, "spread_hc": 1.97,
                      "spread_lc": 1.14, "spread_hl": 3.36, "spread_oc": 1.86},
            "近2周": {"spread_oh": 2.22, "spread_ol": 1.44, "spread_hc": 2.03,
                      "spread_lc": 1.14, "spread_hl": 3.33, "spread_oc": 1.72},
        }
        composite = {
            k: statistics.mean([window_means[w][k] for w in ["历史", "近3月", "近1月", "近2周"]])
            for k in MODEL_SPREAD_KEYS
        }
        table = _build_reference_table(200.0, window_means, composite)
        assert "最高价预测" in table
        assert "最低价预测" in table
        assert "收盘价预测" in table
        assert "203.77" in table
        assert "196.18" in table
        oc_comp = composite["spread_oc"]
        assert f"{200.0 - oc_comp:.2f}" in table
        assert "+" in table
        assert "-" in table

    def test_mean_calculation(self):
        window_means = {
            "历史": {"spread_oh": 4.0, "spread_ol": 4.0, "spread_hc": 4.0,
                     "spread_lc": 4.0, "spread_hl": 4.0, "spread_oc": 4.0},
            "近3月": {"spread_oh": 2.0, "spread_ol": 2.0, "spread_hc": 2.0,
                      "spread_lc": 2.0, "spread_hl": 2.0, "spread_oc": 2.0},
            "近1月": {"spread_oh": 1.0, "spread_ol": 1.0, "spread_hc": 1.0,
                      "spread_lc": 1.0, "spread_hl": 1.0, "spread_oc": 1.0},
            "近2周": {"spread_oh": 0.5, "spread_ol": 0.5, "spread_hc": 0.5,
                      "spread_lc": 0.5, "spread_hl": 0.5, "spread_oc": 0.5},
        }
        composite = {k: 1.875 for k in MODEL_SPREAD_KEYS}
        table = _build_reference_table(100.0, window_means, composite)
        assert "101.88" in table
        assert "98.12" in table


class TestBuildTradingPlan:
    def test_with_data(self):
        rows = [
            {"trade_date": "2024-01-04", "spread_oh": 1.0, "spread_ol": 0.5,
             "spread_hl": 1.5, "spread_oc": 0.5, "spread_hc": 0.5, "spread_lc": 1.0},
            {"trade_date": "2024-01-03", "spread_oh": 2.0, "spread_ol": 1.0,
             "spread_hl": 3.0, "spread_oc": 1.0, "spread_hc": 1.0, "spread_lc": 2.0},
        ]
        plan = build_trading_plan("603778", 200.0, rows)
        assert "603778 交易计划" in plan
        assert "价差模型" in plan
        assert "历史参考价" in plan
        assert "开盘价: 200.00" in plan

    def test_empty_data(self):
        plan = build_trading_plan("603778", 200.0, [])
        assert "暂无历史数据" in plan


