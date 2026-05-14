"""One-shot helper that dumps analysis fixtures for Go parity tests.

Run from project root:

    python tools/dump_python_fixture.py 603778 8.50 8.80 8.40 8.55 > pkg/analysis/testdata/603778.json

Args:
    code               6-digit stock code, present in the local SQLite db
    open               today's open (pass 0 to skip header / plan)
    high low close     optional actual overrides
"""
import json
import sys
from datetime import datetime

from analysis import StockAnalyzer
from config import DB_PATH
from db import DailyDB


def main() -> None:
    if len(sys.argv) < 2:
        print("usage: dump_python_fixture.py CODE [OPEN] [HIGH LOW CLOSE]", file=sys.stderr)
        sys.exit(2)
    code = sys.argv[1]
    open_p = float(sys.argv[2]) if len(sys.argv) > 2 else None
    high = float(sys.argv[3]) if len(sys.argv) > 3 else None
    low = float(sys.argv[4]) if len(sys.argv) > 4 else None
    close = float(sys.argv[5]) if len(sys.argv) > 5 else None

    db = DailyDB(DB_PATH)
    db.init()
    rows = db.query_daily(code, "2000-01-01", datetime.now().strftime("%Y-%m-%d"))

    a = StockAnalyzer(
        code,
        all_rows=rows,
        open_price=open_p,
        actual_high=high,
        actual_low=low,
        actual_close=close,
    )
    window_means = a._compute_window_means()
    composite = a._compute_composite_means(window_means)

    fixture = {
        "code": code,
        "rows": rows,
        "open_price": open_p,
        "actual_high": high,
        "actual_low": low,
        "actual_close": close,
        "window_means": {
            w: {k: window_means[w][k] for k in a.MODEL_SPREAD_KEYS}
            for w in a._WINDOW_NAMES
        },
        "composite_means": composite,
        "header_text": a._format_header(open_p, composite) if open_p is not None else "",
        "model_table_text": a._build_spread_model_table(window_means, composite),
        "reference_table_text": a._build_reference_table(open_p, window_means, composite)
            if open_p is not None else "",
    }
    json.dump(fixture, sys.stdout, ensure_ascii=False, indent=2, default=str)


if __name__ == "__main__":
    main()
