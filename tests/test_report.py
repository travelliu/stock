import statistics
import pytest
from report import _compute_window_means, _compute_composite_means, MODEL_SPREAD_KEYS


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
