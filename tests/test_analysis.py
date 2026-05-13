import pytest
from analysis import StockAnalyzer


@pytest.fixture
def analyzer():
    return StockAnalyzer("603778")


class TestComputeStatistics:
    def test_basic_statistics(self, analyzer):
        rows = [
            {"spread_oh": 1.0, "spread_ol": 0.5, "spread_hl": 1.5,
             "spread_oc": 0.5, "spread_hc": 0.5, "spread_lc": 1.0},
            {"spread_oh": 2.0, "spread_ol": 1.0, "spread_hl": 3.0,
             "spread_oc": 1.0, "spread_hc": 1.0, "spread_lc": 2.0},
            {"spread_oh": 0.5, "spread_ol": 0.25, "spread_hl": 0.75,
             "spread_oc": 0.25, "spread_hc": 0.25, "spread_lc": 0.5},
        ]
        stats = analyzer.compute_statistics(rows)
        assert stats["count"] == 3
        assert stats["spreads"]["spread_oh"]["mean"] == pytest.approx(1.1667, rel=1e-3)
        assert stats["spreads"]["spread_oh"]["median"] == pytest.approx(1.0)

    def test_empty_rows(self, analyzer):
        stats = analyzer.compute_statistics([])
        assert stats["count"] == 0
        assert stats["spreads"] == {}

    def test_single_row(self, analyzer):
        rows = [
            {"spread_oh": 1.0, "spread_ol": 0.5, "spread_hl": 1.5,
             "spread_oc": 0.5, "spread_hc": 0.5, "spread_lc": 1.0},
        ]
        stats = analyzer.compute_statistics(rows)
        assert stats["count"] == 1
        assert stats["spreads"]["spread_oh"]["mean"] == pytest.approx(1.0)
        assert stats["spreads"]["spread_oh"]["median"] == pytest.approx(1.0)


class TestSpreadLabels:
    def test_all_keys_present(self, analyzer):
        expected_keys = {
            "spread_oh", "spread_ol", "spread_hl",
            "spread_oc", "spread_hc", "spread_lc",
        }
        assert set(analyzer.SPREAD_LABELS.keys()) == expected_keys

    def test_labels_are_chinese(self, analyzer):
        for label in analyzer.SPREAD_LABELS.values():
            assert any("\u4e00" <= c <= "\u9fff" for c in label)


class TestComputeDistribution:
    def test_basic_distribution(self, analyzer):
        values = [1.0, 2.0, 3.0, 4.0, 5.0]
        bins = analyzer.compute_distribution(values, num_bins=5)
        assert len(bins) == 5
        total_count = sum(b["count"] for b in bins)
        assert total_count == 5

    def test_empty_values(self, analyzer):
        bins = analyzer.compute_distribution([])
        assert bins == []

    def test_single_value(self, analyzer):
        bins = analyzer.compute_distribution([3.0])
        assert len(bins) == 1
        assert bins[0]["count"] == 1
        assert bins[0]["pct"] == 100.0


class TestComputeRecommendedRange:
    def test_empty_returns_none(self, analyzer):
        result = analyzer.compute_recommended_range([])
        assert result is None

    def test_single_value(self, analyzer):
        result = analyzer.compute_recommended_range([3.0])
        assert result is not None
        assert result["low"] == pytest.approx(3.0)
        assert result["high"] == pytest.approx(3.0)
        assert result["cum_pct"] == pytest.approx(100.0)

    def test_sliding_window_contiguous(self, analyzer):
        values = list(range(10))
        result = analyzer.compute_recommended_range(values, threshold=30.0)
        assert result is not None
        assert result["high"] - result["low"] == pytest.approx(2.0)

    def test_skewed_distribution_tight_range(self, analyzer):
        values = [0.1, 0.2, 0.15, 0.25, 0.3, 0.35, 0.18, 0.22, 1.0, 2.0]
        result = analyzer.compute_recommended_range(values, threshold=60.0)
        assert result is not None
        assert result["cum_pct"] >= 60.0
        assert result["high"] - result["low"] < 1.0

    def test_high_threshold_nearly_full_range(self, analyzer):
        values = list(range(100))
        result = analyzer.compute_recommended_range(values, threshold=95.0)
        assert result is not None
        assert result["high"] - result["low"] == pytest.approx(94.0)

    def test_default_threshold_is_60(self, analyzer):
        values = [0.1] * 40 + [5.0] * 10
        result = analyzer.compute_recommended_range(values)
        assert result is not None
        assert result["low"] == pytest.approx(0.1)
        assert result["high"] == pytest.approx(0.1)
        assert result["cum_pct"] >= 60.0


class TestComputeWindowMeans:
    def test_basic(self, analyzer):
        analyzer.all_rows = [
            {"trade_date": "2024-01-04", "spread_oh": 1.0, "spread_ol": 0.5,
             "spread_hl": 1.5, "spread_oc": 0.5, "spread_hc": 0.5, "spread_lc": 1.0},
            {"trade_date": "2024-01-03", "spread_oh": 2.0, "spread_ol": 1.0,
             "spread_hl": 3.0, "spread_oc": 1.0, "spread_hc": 1.0, "spread_lc": 2.0},
            {"trade_date": "2024-01-02", "spread_oh": 3.0, "spread_ol": 1.5,
             "spread_hl": 4.5, "spread_oc": 1.5, "spread_hc": 1.5, "spread_lc": 3.0},
        ]
        means = analyzer._compute_window_means()
        assert means["历史"]["spread_oh"] == pytest.approx(2.0)
        assert means["历史"]["spread_ol"] == pytest.approx(1.0)
        assert means["近3月"]["spread_oh"] == pytest.approx(2.0)
        assert means["近2周"]["spread_ol"] == pytest.approx(1.0)

    def test_empty_rows(self, analyzer):
        analyzer.all_rows = []
        means = analyzer._compute_window_means()
        for wname in ["历史", "近3月", "近1月", "近2周"]:
            assert all(v is None for v in means[wname].values())

    def test_rows_with_none_values(self, analyzer):
        analyzer.all_rows = [
            {"trade_date": "2024-01-03", "spread_oh": 1.0, "spread_ol": None,
             "spread_hl": 1.5, "spread_oc": None, "spread_hc": 0.5, "spread_lc": 1.0},
            {"trade_date": "2024-01-02", "spread_oh": 2.0, "spread_ol": 1.0,
             "spread_hl": None, "spread_oc": 1.0, "spread_hc": None, "spread_lc": 2.0},
        ]
        means = analyzer._compute_window_means()
        assert means["历史"]["spread_oh"] == pytest.approx(1.5)
        assert means["历史"]["spread_ol"] == pytest.approx(1.0)
        assert means["历史"]["spread_hl"] == pytest.approx(1.5)
        assert means["历史"]["spread_oc"] == pytest.approx(1.0)
        assert means["历史"]["spread_hc"] == pytest.approx(0.5)
        assert means["历史"]["spread_lc"] == pytest.approx(1.5)
        assert means["近3月"]["spread_oh"] == pytest.approx(1.5)
        assert means["近1月"]["spread_oh"] == pytest.approx(1.5)
        assert means["近2周"]["spread_oh"] == pytest.approx(1.5)


class TestComputeCompositeMeans:
    def test_basic(self, analyzer):
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
        composite = analyzer._compute_composite_means(window_means)
        assert composite["spread_oh"] == pytest.approx(1.875)
        assert composite["spread_ol"] == pytest.approx(1.1875)

    def test_with_none_values(self, analyzer):
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
        composite = analyzer._compute_composite_means(window_means)
        assert composite["spread_oh"] == pytest.approx(1.875)
        assert composite["spread_ol"] == 0.0

    def test_all_none_values(self, analyzer):
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
        composite = analyzer._compute_composite_means(window_means)
        assert composite["spread_oh"] == 0.0
        assert composite["spread_ol"] == 0.0


class TestTableFormatting:
    def test_display_width_english(self, analyzer):
        assert analyzer._display_width("hello") == 5

    def test_display_width_chinese(self, analyzer):
        assert analyzer._display_width("开盘") == 4

    def test_display_width_mixed(self, analyzer):
        assert analyzer._display_width("开A") == 3

    def test_format_table_basic(self, analyzer):
        headers = ["时段", "数值"]
        rows = [["历史", "1.23"], ["近1月", "0.45"]]
        table = analyzer._format_table(headers, rows)
        assert "时段" in table
        assert "历史" in table
        assert "+" in table
        lines = table.split("\n")
        assert len(lines) == 6


class TestFormatHeader:
    def test_basic(self, analyzer):
        composite = {"spread_oh": 3.77, "spread_ol": 3.82,
                     "spread_oc": 4.07, "spread_hl": 8.21,
                     "spread_hc": 4.07, "spread_lc": 3.25}
        header = analyzer._format_header(200.0, composite)
        assert "开盘价: 200.00" in header
        assert "最高价: 203.77" in header
        assert "最低价: 196.18" in header
        assert "收盘价: 195.93" in header


class TestBuildSpreadModelTable:
    def test_structure(self, analyzer):
        window_means = {
            "历史": {k: 1.0 for k in analyzer.MODEL_SPREAD_KEYS},
            "近3月": {k: 0.5 for k in analyzer.MODEL_SPREAD_KEYS},
            "近1月": {k: 0.3 for k in analyzer.MODEL_SPREAD_KEYS},
            "近2周": {k: 0.2 for k in analyzer.MODEL_SPREAD_KEYS},
        }
        composite = {k: 0.5 for k in analyzer.MODEL_SPREAD_KEYS}
        table = analyzer._build_spread_model_table(window_means, composite)
        assert "时段" in table
        assert "开盘与最高价" in table
        assert "综合均值" in table
        assert "1.00" in table
        lines = table.split("\n")
        assert len(lines) == 9

    def test_with_none(self, analyzer):
        window_means = {
            "历史": {k: 1.0 for k in analyzer.MODEL_SPREAD_KEYS},
            "近3月": {k: None for k in analyzer.MODEL_SPREAD_KEYS},
            "近1月": {k: None for k in analyzer.MODEL_SPREAD_KEYS},
            "近2周": {k: None for k in analyzer.MODEL_SPREAD_KEYS},
        }
        composite = {k: 1.0 for k in analyzer.MODEL_SPREAD_KEYS}
        table = analyzer._build_spread_model_table(window_means, composite)
        assert "-" in table


class TestBuildReferenceTable:
    def test_high_low_close_rows(self, analyzer):
        import statistics
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
            for k in analyzer.MODEL_SPREAD_KEYS
        }
        table = analyzer._build_reference_table(200.0, window_means, composite)
        assert "最高价预测" in table
        assert "最低价预测" in table
        assert "收盘价预测" in table
        assert "203.77" in table
        assert "196.18" in table
        assert "+" in table
        assert "-" in table

    def test_mean_calculation(self, analyzer):
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
        composite = {k: 1.875 for k in analyzer.MODEL_SPREAD_KEYS}
        table = analyzer._build_reference_table(100.0, window_means, composite)
        assert "101.88" in table
        assert "98.12" in table


class TestBuildTradingPlan:
    def test_with_data(self, analyzer):
        analyzer.all_rows = [
            {"trade_date": "2024-01-04", "spread_oh": 1.0, "spread_ol": 0.5,
             "spread_hl": 1.5, "spread_oc": 0.5, "spread_hc": 0.5, "spread_lc": 1.0},
            {"trade_date": "2024-01-03", "spread_oh": 2.0, "spread_ol": 1.0,
             "spread_hl": 3.0, "spread_oc": 1.0, "spread_hc": 1.0, "spread_lc": 2.0},
        ]
        analyzer.open_price = 200.0
        plan = analyzer.build_trading_plan()
        assert "603778 交易计划" in plan
        assert "价差模型" in plan
        assert "历史参考价" in plan
        assert "开盘价: 200.00" in plan

    def test_empty_data(self, analyzer):
        analyzer.all_rows = []
        analyzer.open_price = 200.0
        plan = analyzer.build_trading_plan()
        assert "暂无历史数据" in plan
