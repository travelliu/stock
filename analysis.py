import statistics
import unicodedata
from typing import Any


class StockAnalyzer:
    """Stock spread analysis and trading plan generator."""

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

    _WINDOW_DAYS = [None, 90, 30, 15]
    _WINDOW_NAMES = ["历史", "近3月", "近1月", "近2周"]

    MODEL_SPREAD_KEYS = [
        "spread_oh", "spread_ol", "spread_hl",
        "spread_hc", "spread_lc", "spread_oc",
    ]

    MODEL_SPREAD_LABELS = [
        "开盘与最高价", "开盘与最低价", "最高与最低价",
        "最高与收盘价", "最低与收盘价", "开盘与收盘价",
    ]

    def __init__(
        self,
        stock: str,
        all_rows: list[dict] | None = None,
        start_date: str = "",
        end_date: str = "",
        show_all: bool = False,
        open_price: float | None = None,
        actual_low: float | None = None,
        actual_high: float | None = None,
    ):
        self.stock = stock
        self.all_rows = all_rows or []
        self.start_date = start_date
        self.end_date = end_date
        self.show_all = show_all
        self.open_price = open_price
        self.actual_low = actual_low
        self.actual_high = actual_high
        self._label: str | None = None

    @property
    def label(self) -> str:
        if self._label is None:
            from company import get_stock_name
            name = get_stock_name(self.stock)
            self._label = f"{self.stock} {name}" if name else self.stock
        return self._label

    # -- Core analysis --

    def make_windows(self, rows_sorted: list[dict]) -> list[tuple[str, list[dict]]]:
        """Create time-window slices from sorted rows.

        Args:
            rows_sorted: rows sorted by trade_date descending

        Returns:
            list of (window_name, sliced_rows)
        """
        return [
            (name, rows_sorted if days is None else rows_sorted[:days])
            for name, days in zip(self._WINDOW_NAMES, self._WINDOW_DAYS)
        ]

    def compute_statistics(
        self,
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
        for key in self.SPREAD_KEYS:
            values = [r[key] for r in rows if r.get(key) is not None]
            if values:
                spreads[key] = {
                    "mean": statistics.mean(values),
                    "median": statistics.median(values),
                }

        return {"count": len(rows), "spreads": spreads}

    def compute_distribution(
        self,
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

        if v_min == v_max:
            return [{"low": v_min, "high": v_max,
                     "count": len(values), "pct": 100.0}]

        bin_width = (v_max - v_min) / num_bins
        bins = []
        for i in range(num_bins):
            low = v_min + i * bin_width
            high = v_min + (i + 1) * bin_width
            count = sum(1 for v in values if low <= v < high)
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

    def compute_recommended_range(
        self,
        values: list[float],
        threshold: float = 60.0,
    ) -> dict[str, Any] | None:
        """Return the narrowest contiguous range covering >= threshold% of observations.

        Uses a sliding window on sorted values to find the tightest interval.

        Args:
            values: raw spread values for one window
            threshold: minimum cumulative percentage (default 60%)

        Returns:
            {"low": float, "high": float, "cum_pct": float} or None if no data
        """
        if not values:
            return None

        sorted_vals = sorted(values)
        n = len(sorted_vals)

        if n == 1:
            return {"low": sorted_vals[0], "high": sorted_vals[0], "cum_pct": 100.0}

        needed = max(1, round(n * threshold / 100))

        best_low = sorted_vals[0]
        best_high = sorted_vals[-1]
        best_span = best_high - best_low

        for i in range(n - needed + 1):
            span = sorted_vals[i + needed - 1] - sorted_vals[i]
            if span < best_span:
                best_span = span
                best_low = sorted_vals[i]
                best_high = sorted_vals[i + needed - 1]

        cum_pct = round(needed / n * 100, 1)
        return {"low": best_low, "high": best_high, "cum_pct": cum_pct}

    # -- Formatting helpers --

    def _display_width(self, s: str) -> int:
        """Calculate terminal display width (CJK chars = 2 columns)."""
        width = 0
        for ch in str(s):
            eaw = unicodedata.east_asian_width(ch)
            width += 2 if eaw in ("W", "F") else 1
        return width

    def _rpad(self, s: str, width: int) -> str:
        """Right-pad string to reach target display width."""
        return str(s) + " " * max(0, width - self._display_width(s))

    def _lpad(self, s: str, width: int) -> str:
        """Left-pad string to reach target display width."""
        return " " * max(0, width - self._display_width(s)) + str(s)

    def _format_table(self, headers: list[str], rows: list[list[str]]) -> str:
        """Format a table with CJK-aware column alignment."""
        col_widths = [self._display_width(h) for h in headers]
        for row in rows:
            for i, cell in enumerate(row):
                if i < len(col_widths):
                    col_widths[i] = max(col_widths[i], self._display_width(cell))
        sep = "+" + "+".join("-" * (w + 2) for w in col_widths) + "+"
        lines = [sep]
        header_line = "|"
        for i, h in enumerate(headers):
            header_line += " " + self._rpad(h, col_widths[i]) + " |"
        lines.append(header_line)
        lines.append(sep)
        for row in rows:
            data_line = "|"
            for i, cell in enumerate(row):
                if i < len(col_widths):
                    data_line += " " + self._rpad(cell, col_widths[i]) + " |"
            lines.append(data_line)
        lines.append(sep)
        return "\n".join(lines)

    def _join_tables_side_by_side(self, tables: list[str], gaps: int = 4) -> str:
        """Join multiple table strings side by side with CJK-aware alignment."""
        if not tables:
            return ""
        split = [t.split("\n") for t in tables]

        normalized = []
        for block in split:
            if not block:
                continue
            max_w = max(self._display_width(line) for line in block)
            padded = [self._rpad(line, max_w) for line in block]
            normalized.append(padded)

        max_lines = max(len(b) for b in normalized)
        pad = " " * gaps
        lines = []
        for i in range(max_lines):
            parts = []
            for b in normalized:
                parts.append(b[i] if i < len(b) else " " * self._display_width(b[0]))
            lines.append(pad.join(parts))
        return "\n".join(lines)

    # -- Trading plan computation --

    def _compute_window_means(self) -> dict[str, dict[str, float | None]]:
        """Compute mean for each spread key in each time window."""
        all_rows_sorted = sorted(self.all_rows, key=lambda r: r["trade_date"], reverse=True)
        windows = self.make_windows(all_rows_sorted)
        result: dict[str, dict[str, float | None]] = {}
        for wname, rows in windows:
            result[wname] = {}
            for key in self.MODEL_SPREAD_KEYS:
                values = [r[key] for r in rows if r.get(key) is not None]
                result[wname][key] = statistics.mean(values) if values else None
        return result

    def _compute_composite_means(
        self, window_means: dict[str, dict[str, float | None]]
    ) -> dict[str, float]:
        """Arithmetic average across the four window means per spread key."""
        composite: dict[str, float] = {}
        for key in self.MODEL_SPREAD_KEYS:
            vals = [
                window_means[w][key]
                for w in self._WINDOW_NAMES
                if window_means[w].get(key) is not None
            ]
            composite[key] = statistics.mean(vals) if vals else 0.0
        return composite

    def _format_header(
        self, open_price: float, composite_means: dict[str, float]
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
        self,
        window_means: dict[str, dict[str, float | None]],
        composite_means: dict[str, float],
    ) -> str:
        headers = ["时段"] + self.MODEL_SPREAD_LABELS
        rows: list[list[str]] = []
        for wname in self._WINDOW_NAMES:
            row = [wname]
            for key in self.MODEL_SPREAD_KEYS:
                val = window_means[wname].get(key)
                row.append(f"{val:.2f}" if val is not None else "-")
            rows.append(row)
        comp_row = ["综合均值"]
        for key in self.MODEL_SPREAD_KEYS:
            val = composite_means.get(key, 0.0)
            comp_row.append(f"{val:.2f}")
        rows.append(comp_row)
        return self._format_table(headers, rows)

    def _build_reference_table(
        self,
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

        high_row = ["最高价预测"]
        for wname in self._WINDOW_NAMES:
            val = window_means[wname].get("spread_oh")
            high_row.append(
                f"{open_price + val:.2f}" if val is not None else "/"
            )
        lc_hist = window_means["历史"].get("spread_lc")
        if self.actual_low is not None and lc_hist is not None:
            high_row.append(f"{self.actual_low + lc_hist:.2f}")
        else:
            high_row.append("/")
        if self.actual_high is not None and lc_hist is not None:
            high_row.append(f"{self.actual_high - lc_hist:.2f}")
        else:
            high_row.append("/")
        nums = []
        for cell in high_row[1:5]:
            try:
                nums.append(float(cell))
            except ValueError:
                pass
        high_row.append(
            f"{statistics.mean(nums):.2f}" if nums else "/"
        )
        high_row.append("+")
        rows.append(high_row)

        low_row = ["最低价预测"]
        for wname in self._WINDOW_NAMES:
            val = window_means[wname].get("spread_ol")
            low_row.append(
                f"{open_price - val:.2f}" if val is not None else "/"
            )
        low_row.append("/")
        hc_hist = window_means["历史"].get("spread_hc")
        if self.actual_high is not None and hc_hist is not None:
            low_row.append(f"{self.actual_high - hc_hist:.2f}")
        else:
            low_row.append("/")
        nums = []
        for cell in low_row[1:5]:
            try:
                nums.append(float(cell))
            except ValueError:
                pass
        low_row.append(
            f"{statistics.mean(nums):.2f}" if nums else "/"
        )
        low_row.append("-")
        rows.append(low_row)

        close_row = ["收盘价预测"]
        oc_comp = composite_means.get("spread_oc", 0.0)
        close_val = open_price - oc_comp
        close_row.append(f"{close_val:.2f}")
        close_row.extend(["/", "/", "/"])
        close_row.extend(["/", "/"])
        close_row.append(f"{close_val:.2f}")
        close_row.append("-")
        rows.append(close_row)

        return self._format_table(headers, rows)

    def build_trading_plan(self) -> str:
        """Build a trading plan report for the given stock."""
        if not self.all_rows:
            return "暂无历史数据，无法生成交易计划。"

        window_means = self._compute_window_means()
        composite_means = self._compute_composite_means(window_means)

        lines = [
            f"=== {self.stock} 交易计划 ===",
            "",
            self._format_header(self.open_price, composite_means),
            "",
            "── 价差模型 ──",
            self._build_spread_model_table(window_means, composite_means),
            "",
            "── 历史参考价 ──",
            self._build_reference_table(self.open_price, window_means, composite_means),
        ]
        return "\n".join(lines)

    # -- Analysis print --

    def print_analysis(self) -> None:
        """Print spread analysis tables for the given stock."""
        all_rows_sorted = sorted(
            self.all_rows, key=lambda r: r["trade_date"], reverse=True
        )
        windows = self.make_windows(all_rows_sorted)

        spread_keys = self.SPREAD_KEYS if self.show_all else self.DEFAULT_SPREADS

        print(f"=== 价差分析 ===")
        print()

        for i in range(0, len(spread_keys), 2):
            pair = spread_keys[i : i + 2]

            if not self.show_all:
                u_headers = [
                    "时段", "样本数", "均值", "中位数", "众数", "",
                    "样本数", "均值", "中位数", "众数", "",
                    "高抛差价(高-开盘)", "低吸差价(开盘-低)",
                ]
                ordered_windows = list(reversed(windows))
                u_table: list[list[str]] = []
                for wname, rows in ordered_windows:
                    oh_vals = [
                        r["spread_oh"] for r in rows
                        if r.get("spread_oh") is not None
                    ]
                    ol_vals = [
                        r["spread_ol"] for r in rows
                        if r.get("spread_ol") is not None
                    ]

                    row: list[str] = [wname]

                    if oh_vals:
                        try:
                            mode_val = statistics.mode(oh_vals)
                        except statistics.StatisticsError:
                            mode_val = "-"
                        row.extend([
                            str(len(oh_vals)),
                            f"{statistics.mean(oh_vals):.2f}",
                            f"{statistics.median(oh_vals):.2f}",
                            f"{mode_val:.2f}"
                            if isinstance(mode_val, float)
                            else str(mode_val),
                        ])
                    else:
                        row.extend(["0", "-", "-", "-"])

                    row.append("")

                    if ol_vals:
                        try:
                            mode_val = statistics.mode(ol_vals)
                        except statistics.StatisticsError:
                            mode_val = "-"
                        row.extend([
                            str(len(ol_vals)),
                            f"{statistics.mean(ol_vals):.2f}",
                            f"{statistics.median(ol_vals):.2f}",
                            f"{mode_val:.2f}"
                            if isinstance(mode_val, float)
                            else str(mode_val),
                        ])
                    else:
                        row.extend(["0", "-", "-", "-"])

                    row.append("")

                    rec_threshold = 60.0 if wname == "全部" else 30.0
                    oh_range = (
                        self.compute_recommended_range(
                            oh_vals, threshold=rec_threshold
                        )
                        if oh_vals else None
                    )
                    ol_range = (
                        self.compute_recommended_range(
                            ol_vals, threshold=rec_threshold
                        )
                        if ol_vals else None
                    )
                    row.append(
                        f"{oh_range['low']:.2f}~{oh_range['high']:.2f} "
                        f"({oh_range['cum_pct']:.1f}%)"
                        if oh_range else "-"
                    )
                    row.append(
                        f"{ol_range['low']:.2f}~{ol_range['high']:.2f} "
                        f"({ol_range['cum_pct']:.1f}%)"
                        if ol_range else "-"
                    )

                    u_table.append(row)

                col_widths = [self._display_width(h) for h in u_headers]
                for urow in u_table:
                    for ci, cell in enumerate(urow):
                        col_widths[ci] = max(
                            col_widths[ci], self._display_width(cell)
                        )

                def _section_w(start: int, end: int) -> int:
                    return sum(col_widths[start:end + 1]) + 3 * (end - start + 1)

                time_sw = col_widths[0] + 2
                oh_sw = _section_w(1, 5)
                ol_sw = _section_w(6, 10)
                rec_sw = _section_w(11, 12)

                sub_line = (
                    "|" + " " * time_sw + "|"
                    + self._rpad("── 最高-开盘 ──", oh_sw) + "|"
                    + self._rpad("── 开盘-最低 ──", ol_sw) + "|"
                    + self._rpad(
                        "── 高抛低吸推荐 (累计占比) ──", rec_sw
                    ) + "|"
                )
                print(sub_line)
                print(self._format_table(u_headers, u_table))
                print()
            else:
                summaries = []
                for key in pair:
                    label = self.SPREAD_LABELS[key]
                    headers = ["时段", "样本数", "均值", "中位数", "众数"]
                    table = []
                    for wname, rows in windows:
                        values = [
                            r[key] for r in rows if r.get(key) is not None
                        ]
                        if values:
                            try:
                                mode_val = statistics.mode(values)
                            except statistics.StatisticsError:
                                mode_val = "-"
                            table.append([
                                wname,
                                str(len(values)),
                                f"{statistics.mean(values):.2f}",
                                f"{statistics.median(values):.2f}",
                                f"{mode_val:.2f}"
                                if isinstance(mode_val, float)
                                else str(mode_val),
                            ])
                        else:
                            table.append([wname, "0", "-", "-"])
                    summaries.append(
                        f"── {label} 汇总 ──\n"
                        + self._format_table(headers, table)
                    )
                print(self._join_tables_side_by_side(summaries))
                print()

            for key in pair:
                label = self.SPREAD_LABELS[key]
                dist_tables = []
                for wname, rows in windows:
                    values = [
                        r[key] for r in rows if r.get(key) is not None
                    ]
                    if not values:
                        continue
                    bins = self.compute_distribution(values)
                    dist_headers = ["区间", "数量", "占比"]
                    dist_table = []
                    for b in bins:
                        interval = f"{b['low']:.2f}~{b['high']:.2f}"
                        dist_table.append([
                            interval,
                            str(b["count"]),
                            f"{b['pct']:.1f}%",
                        ])
                    dist_tables.append(
                        f"── {label} 分布 ({wname},{len(values)}条) ──\n"
                        + self._format_table(dist_headers, dist_table)
                    )
                if dist_tables:
                    print(self._join_tables_side_by_side(dist_tables))
            print()

    def print_table(self, rows: list[dict]) -> None:
        """Print daily data table."""
        headers = [
            "trade_date", "open", "high", "low", "close",
            "vol", "spread_oh", "spread_ol", "spread_hl",
            "spread_oc", "spread_hc", "spread_lc",
        ]
        display_names = [
            "日期", "开盘", "最高", "最低", "收盘",
            "成交量(万)", "高-开", "开-低", "高-低",
            "开-收", "高-收", "低-收",
        ]
        table = []
        for r in reversed(rows):
            row = []
            for h in headers:
                v = r.get(h, "")
                if h == "vol":
                    row.append(f"{v / 10000:>8.2f}")
                elif isinstance(v, float):
                    row.append(f"{v:>6.2f}")
                else:
                    row.append(str(v))
            table.append(row)

        print(self._format_table(display_names, table))

    def show(self) -> None:
        """Show trading plan, spread analysis, and daily data table."""

        print(f"=== {self.label} ===")

        if self.open_price is not None:
            plan = self.build_trading_plan()
            print(plan)
            print()

        if self.all_rows:
            self.print_analysis()

        rows = [
            r for r in self.all_rows
            if self.start_date <= r["trade_date"] <= self.end_date
        ] if self.all_rows else []

        if rows:
            print(
                f"=== 日线数据 === "
                f"({self.start_date} ~ {self.end_date}) 共 {len(rows)} 条 ---"
            )
            print()
            self.print_table(rows)
        elif not self.all_rows:
            print(f"No data for {self.stock}")
