import statistics
from typing import Any

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
        for key in MODEL_SPREAD_KEYS:
            values = [r[key] for r in rows if r.get(key) is not None]
            result[wname][key] = statistics.mean(values) if values else None
    return result


def _compute_composite_means(
    window_means: dict[str, dict[str, float | None]],
) -> dict[str, float]:
    """Arithmetic average across the four window means per spread key."""
    composite: dict[str, float] = {}
    for key in MODEL_SPREAD_KEYS:
        vals = [
            window_means[w][key]
            for w in _WINDOW_NAMES
            if window_means[w].get(key) is not None
        ]
        composite[key] = statistics.mean(vals) if vals else 0.0
    return composite


import unicodedata


def _display_width(s: str) -> int:
    width = 0
    for ch in str(s):
        eaw = unicodedata.east_asian_width(ch)
        width += 2 if eaw in ("W", "F") else 1
    return width


def _rpad(s: str, width: int) -> str:
    return str(s) + " " * max(0, width - _display_width(s))


def _format_table(headers: list[str], rows: list[list[str]]) -> str:
    col_widths = [_display_width(h) for h in headers]
    for row in rows:
        for i, cell in enumerate(row):
            if i < len(col_widths):
                col_widths[i] = max(col_widths[i], _display_width(cell))
    sep = "+" + "+".join("-" * (w + 2) for w in col_widths) + "+"
    lines = [sep]
    header_line = "|"
    for i, h in enumerate(headers):
        header_line += " " + _rpad(h, col_widths[i]) + " |"
    lines.append(header_line)
    lines.append(sep)
    for row in rows:
        data_line = "|"
        for i, cell in enumerate(row):
            if i < len(col_widths):
                data_line += " " + _rpad(cell, col_widths[i]) + " |"
        lines.append(data_line)
    lines.append(sep)
    return "\n".join(lines)
