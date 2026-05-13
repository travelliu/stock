#!/bin/bash
# Stock daily data fetcher + analysis report generator
# Mon-Fri 22:00 (after market data settled)
#
# Install crontab:
#   crontab scripts/fetch_daily.cron
#
# Or manually add:
#   0 22 * * 1-5 /root/code/github/travelliu/stock/scripts/fetch_daily.sh >> /root/code/github/travelliu/stock/data/fetch.log 2>&1

# set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PYTHON="$PROJECT_DIR/.venv/bin/python"
REPORTS_DIR="$PROJECT_DIR/reports"

cd "$PROJECT_DIR"

# ── Step 1: Fetch latest daily data ──
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Fetching daily data ..."
"$PYTHON" stock.py fetch

# ── Step 2: Generate analysis report via Claude ──
mkdir -p "$REPORTS_DIR"

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Generating analysis report via Claude ..."

claude --print \
    "分析 DEFAULT_STOCKS 里所有股票的价差数据，生成综合做T分析报告（参考 stock_analysis_20260513.md 的格式），包含综合速览表、逐股详细分析、做T优先级排序、明日操作要点。
判断今天之后最近的下一个A股交易日，报告文件写到 ${REPORTS_DIR}/stock_analysis_下一个交易日YYYYMMDD.md。
如果同名文件已存在则跳过不写。" \
    > /dev/null

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Done."
