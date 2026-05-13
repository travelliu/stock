#!/bin/bash
# Stock daily data fetcher - run as crontab job
# Mon-Fri 22:00 (after market data settled)
#
# Install crontab:
#   crontab scripts/fetch_daily.cron
#
# Or manually add:
#   0 22 * * 1-5 /root/code/github/travelliu/stock/scripts/fetch_daily.sh >> /root/code/github/travelliu/stock/data/fetch.log 2>&1

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PYTHON="$PROJECT_DIR/.venv/bin/python"

cd "$PROJECT_DIR"
"$PYTHON" stock.py fetch
