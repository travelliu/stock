import statistics
from typing import Any

from analysis import SPREAD_KEYS


MODEL_SPREAD_KEYS = [
    "spread_oh", "spread_ol", "spread_hl",
    "spread_hc", "spread_lc", "spread_oc",
]
MODEL_SPREAD_LABELS = [
    "开盘与最高价", "开盘与最低价", "最高与最低价",
    "最高与收盘价", "最低与收盘价", "开盘与收盘价",
]
_WINDOW_NAMES = ["历史", "近3月", "近1月", "近2周"]


def _compute_window_means(
    all_rows: list[dict[str, Any]],
) -> dict[str, dict[str, float | None]]:
    """Compute mean for each spread key in each time window."""
    all_rows_sorted = sorted(all_rows, key=lambda r: r["trade_date"], reverse=True)
    windows = [
        ("历史", all_rows_sorted),
        ("近3月", all_rows_sorted[:90]),
        ("近1月", all_rows_sorted[:30]),
        ("近2周", all_rows_sorted[:15]),
    ]
    result: dict[str, dict[str, float | None]] = {}
    for wname, rows in windows:
        result[wname] = {}
        for key in SPREAD_KEYS:
            values = [r[key] for r in rows if r.get(key) is not None]
            result[wname][key] = statistics.mean(values) if values else None
    return result


def _compute_composite_means(
    window_means: dict[str, dict[str, float | None]],
) -> dict[str, float]:
    """Arithmetic average across the four window means per spread key."""
    composite: dict[str, float] = {}
    for key in SPREAD_KEYS:
        vals = [
            window_means[w][key]
            for w in _WINDOW_NAMES
            if window_means[w].get(key) is not None
        ]
        composite[key] = statistics.mean(vals) if vals else 0.0
    return composite
