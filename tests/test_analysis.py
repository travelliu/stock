import pytest
from analysis import compute_statistics, compute_distribution, compute_recommended_range, SPREAD_LABELS


class TestComputeStatistics:
    def test_basic_statistics(self):
        rows = [
            {"spread_oh": 1.0, "spread_ol": 0.5, "spread_hl": 1.5,
             "spread_oc": 0.5, "spread_hc": 0.5, "spread_lc": 1.0},
            {"spread_oh": 2.0, "spread_ol": 1.0, "spread_hl": 3.0,
             "spread_oc": 1.0, "spread_hc": 1.0, "spread_lc": 2.0},
            {"spread_oh": 0.5, "spread_ol": 0.25, "spread_hl": 0.75,
             "spread_oc": 0.25, "spread_hc": 0.25, "spread_lc": 0.5},
        ]
        stats = compute_statistics(rows)
        assert stats["count"] == 3
        assert stats["spreads"]["spread_oh"]["mean"] == pytest.approx(1.1667, rel=1e-3)
        assert stats["spreads"]["spread_oh"]["median"] == pytest.approx(1.0)

    def test_empty_rows(self):
        stats = compute_statistics([])
        assert stats["count"] == 0
        assert stats["spreads"] == {}

    def test_single_row(self):
        rows = [
            {"spread_oh": 1.0, "spread_ol": 0.5, "spread_hl": 1.5,
             "spread_oc": 0.5, "spread_hc": 0.5, "spread_lc": 1.0},
        ]
        stats = compute_statistics(rows)
        assert stats["count"] == 1
        assert stats["spreads"]["spread_oh"]["mean"] == pytest.approx(1.0)
        assert stats["spreads"]["spread_oh"]["median"] == pytest.approx(1.0)


class TestSpreadLabels:
    def test_all_keys_present(self):
        expected_keys = {
            "spread_oh", "spread_ol", "spread_hl",
            "spread_oc", "spread_hc", "spread_lc",
        }
        assert set(SPREAD_LABELS.keys()) == expected_keys

    def test_labels_are_chinese(self):
        for label in SPREAD_LABELS.values():
            assert any("\u4e00" <= c <= "\u9fff" for c in label)


class TestComputeDistribution:
    def test_basic_distribution(self):
        values = [1.0, 2.0, 3.0, 4.0, 5.0]
        bins = compute_distribution(values, num_bins=5)
        assert len(bins) == 5
        total_count = sum(b["count"] for b in bins)
        assert total_count == 5

    def test_empty_values(self):
        bins = compute_distribution([])
        assert bins == []

    def test_single_value(self):
        bins = compute_distribution([3.0])
        assert len(bins) == 1
        assert bins[0]["count"] == 1
        assert bins[0]["pct"] == 100.0


class TestComputeRecommendedRange:
    def test_empty_returns_none(self):
        result = compute_recommended_range([])
        assert result is None

    def test_single_value(self):
        result = compute_recommended_range([3.0])
        assert result is not None
        assert result["low"] == pytest.approx(3.0)
        assert result["high"] == pytest.approx(3.0)
        assert result["cum_pct"] == pytest.approx(100.0)

    def test_basic_cumulative(self):
        # 10 values from 0.0 to 9.0, 10 bins -> 1 value per bin, each 10%
        values = list(range(10))
        result = compute_recommended_range(values, threshold=30.0)
        assert result is not None
        assert result["cum_pct"] >= 30.0
        # Top 3 bins by density are all tied at 10%, so 3 bins get selected = 30%
        assert result["cum_pct"] == pytest.approx(30.0)

    def test_skewed_distribution(self):
        # Values concentrated in lower range
        values = [0.1, 0.2, 0.15, 0.25, 0.3, 1.0, 2.0]
        result = compute_recommended_range(values, threshold=60.0)
        assert result is not None
        assert result["cum_pct"] >= 60.0
        # Range should cover the lower cluster
        assert result["low"] <= 0.1

    def test_threshold_exceeds_all_bins(self):
        # All bins equally distributed, threshold 95% with 10 bins
        values = list(range(100))
        result = compute_recommended_range(values, threshold=95.0)
        assert result is not None
        # Should include all bins (100%)
        assert result["cum_pct"] == pytest.approx(100.0)

    def test_default_threshold_is_60(self):
        # Create values where top 60%+ falls in a specific range
        values = [0.1] * 40 + [5.0] * 10
        result = compute_recommended_range(values)
        assert result is not None
        assert result["cum_pct"] >= 60.0
