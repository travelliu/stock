import statistics
from typing import Any

SPREAD_KEYS = [
    "spread_oh", "spread_ol", "spread_hl",
    "spread_oc", "spread_hc", "spread_lc",
]

SPREAD_LABELS: dict[str, str] = {
    "spread_oh": "最高-开盘",
    "spread_ol": "开盘-最低",
    "spread_hl": "最高-最低",
    "spread_oc": "开盘-收盘",
    "spread_hc": "最高-收盘",
    "spread_lc": "最低-收盘",
}

DEFAULT_SPREADS = ["spread_oh", "spread_ol"]

NUM_BINS = 10


def compute_statistics(
    rows: list[dict[str, Any]],
) -> dict[str, Any]:
    """Calculate mean and median for each spread type.

    Args:
        rows: list of row dicts from db.query_daily()

    Returns:
        {"count": N, "spreads": {"spread_oh": {"mean": X, "median": Y}, ...}}
    """
    if not rows:
        return {"count": 0, "spreads": {}}

    spreads: dict[str, dict[str, float]] = {}
    for key in SPREAD_KEYS:
        values = [r[key] for r in rows if r.get(key) is not None]
        if values:
            spreads[key] = {
                "mean": statistics.mean(values),
                "median": statistics.median(values),
            }

    return {"count": len(rows), "spreads": spreads}


def compute_distribution(
    values: list[float],
    num_bins: int = NUM_BINS,
) -> list[dict[str, Any]]:
    """Compute histogram distribution for a list of values.

    Args:
        values: list of numeric values
        num_bins: number of equal-width bins

    Returns:
        list of {"low": float, "high": float, "count": int, "pct": float}
    """
    if not values:
        return []

    v_min = min(values)
    v_max = max(values)

    # Handle case where all values are the same
    if v_min == v_max:
        return [{"low": v_min, "high": v_max,
                 "count": len(values), "pct": 100.0}]

    bin_width = (v_max - v_min) / num_bins
    bins = []
    for i in range(num_bins):
        low = v_min + i * bin_width
        high = v_min + (i + 1) * bin_width
        count = sum(1 for v in values if low <= v < high)
        # Last bin includes the upper bound
        if i == num_bins - 1:
            count = sum(1 for v in values if low <= v <= high)
        pct = count / len(values) * 100
        bins.append({
            "low": low,
            "high": high,
            "count": count,
            "pct": round(pct, 1),
        })

    return bins
