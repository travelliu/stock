import pytest
from analysis import compute_statistics, SPREAD_LABELS


class TestComputeStatistics:
    def test_basic_statistics(self):
        rows = [
            {"spread_oh": 1.0, "spread_ol": 0.5, "spread_hl": 1.5,
             "spread_oc": -0.5, "spread_hc": 0.5, "spread_lc": -1.0},
            {"spread_oh": 2.0, "spread_ol": 1.0, "spread_hl": 3.0,
             "spread_oc": -1.0, "spread_hc": 1.0, "spread_lc": -2.0},
            {"spread_oh": 0.5, "spread_ol": 0.25, "spread_hl": 0.75,
             "spread_oc": -0.25, "spread_hc": 0.25, "spread_lc": -0.5},
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
             "spread_oc": -0.5, "spread_hc": 0.5, "spread_lc": -1.0},
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
