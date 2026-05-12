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
