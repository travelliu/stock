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


def _format_header(
    open_price: float,
    composite_means: dict[str, float],
) -> str:
    high_price = open_price + composite_means.get("spread_oh", 0.0)
    low_price = open_price - composite_means.get("spread_ol", 0.0)
    close_price = open_price - composite_means.get("spread_oc", 0.0)
    return (
        f"开盘价: {open_price:.2f}   "
        f"最高价: {high_price:.2f}   "
        f"最低价: {low_price:.2f}   "
        f"收盘价: {close_price:.2f}"
    )


def _build_spread_model_table(
    window_means: dict[str, dict[str, float | None]],
    composite_means: dict[str, float],
) -> str:
    headers = ["时段"] + MODEL_SPREAD_LABELS
    rows: list[list[str]] = []
    for wname in _WINDOW_NAMES:
        row = [wname]
        for key in MODEL_SPREAD_KEYS:
            val = window_means[wname].get(key)
            row.append(f"{val:.2f}" if val is not None else "-")
        rows.append(row)
    comp_row = ["综合均值"]
    for key in MODEL_SPREAD_KEYS:
        val = composite_means.get(key, 0.0)
        comp_row.append(f"{val:.2f}")
    rows.append(comp_row)
    return _format_table(headers, rows)


def _build_reference_table(
    open_price: float,
    window_means: dict[str, dict[str, float | None]],
    composite_means: dict[str, float],
) -> str:
    headers = [
        "", "历史参考价", "近3月参考价", "近1月参考价", "近2周参考价",
        "最低价反推", "最高价反推", "均值", "正负算一",
    ]

    pred_high = open_price + composite_means.get("spread_oh", 0.0)
    pred_low = open_price - composite_means.get("spread_ol", 0.0)
    rows: list[list[str]] = []

    # 最高价预测行
    high_row = ["最高价预测"]
    for wname in _WINDOW_NAMES:
        val = window_means[wname].get("spread_oh")
        high_row.append(f"{open_price + val:.2f}" if val is not None else "/")
    lc_hist = window_means["历史"].get("spread_lc")
    high_row.append(f"{pred_low + lc_hist:.2f}" if lc_hist is not None else "/")
    high_row.append("/")
    nums = []
    for cell in high_row[1:5]:
        try:
            nums.append(float(cell))
        except ValueError:
            pass
    high_row.append(f"{statistics.mean(nums):.2f}" if nums else "/")
    high_row.append("+")
    rows.append(high_row)

    # 最低价预测行
    low_row = ["最低价预测"]
    for wname in _WINDOW_NAMES:
        val = window_means[wname].get("spread_ol")
        low_row.append(f"{open_price - val:.2f}" if val is not None else "/")
    low_row.append("/")
    hc_hist = window_means["历史"].get("spread_hc")
    low_row.append(f"{pred_high - hc_hist:.2f}" if hc_hist is not None else "/")
    nums = []
    for cell in low_row[1:5]:
        try:
            nums.append(float(cell))
        except ValueError:
            pass
    low_row.append(f"{statistics.mean(nums):.2f}" if nums else "/")
    low_row.append("-")
    rows.append(low_row)

    # 收盘价预测行
    close_row = ["收盘价预测"]
    oc_comp = composite_means.get("spread_oc", 0.0)
    close_val = open_price - oc_comp
    close_row.append(f"{close_val:.2f}")
    close_row.extend(["/", "/", "/"])
    close_row.extend(["/", "/"])
    close_row.append(f"{close_val:.2f}")
    close_row.append("-")
    rows.append(close_row)

    return _format_table(headers, rows)


def build_trading_plan(
    stock: str,
    open_price: float,
    all_rows: list[dict[str, Any]],
) -> str:
    if not all_rows:
        return "暂无历史数据，无法生成交易计划。"

    window_means = _compute_window_means(all_rows)
    composite_means = _compute_composite_means(window_means)

    lines = [
        f"=== {stock} 交易计划 ===",
        "",
        _format_header(open_price, composite_means),
        "",
        "── 价差模型 ──",
        _build_spread_model_table(window_means, composite_means),
        "",
        "── 历史参考价 ──",
        _build_reference_table(open_price, window_means, composite_means),
    ]
    return "\n".join(lines)
