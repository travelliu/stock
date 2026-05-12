import subprocess
import sys


def run_cli(*args: str) -> subprocess.CompletedProcess:
    return subprocess.run(
        [sys.executable, "stock.py", *args],
        capture_output=True,
        text=True,
        cwd="/root/code/github/travelliu/stock",
    )


class TestCLIHelp:
    def test_no_args_shows_help(self):
        result = run_cli()
        assert result.returncode != 0 or "usage" in result.stdout.lower() or "usage" in result.stderr.lower()

    def test_fetch_help(self):
        result = run_cli("fetch", "--help")
        assert result.returncode == 0
        assert "--stocks" in result.stdout
        assert "--days" in result.stdout

    def test_show_help(self):
        result = run_cli("show", "--help")
        assert result.returncode == 0
        assert "--stock" in result.stdout
        assert "--analyze" in result.stdout


class TestCLIShow:
    def test_show_requires_stock(self):
        result = run_cli("show")
        assert result.returncode != 0
