import pytest
from config import to_tushare_code, TUSHARE_TOKEN, DB_PATH, DEFAULT_STOCKS


class TestToTushareCode:
    # Shanghai main board
    def test_shanghai_600(self):
        assert to_tushare_code("600537") == "600537.SH"

    def test_shanghai_601(self):
        assert to_tushare_code("601398") == "601398.SH"

    def test_shanghai_603(self):
        assert to_tushare_code("603778") == "603778.SH"

    def test_shanghai_605(self):
        assert to_tushare_code("605123") == "605123.SH"

    # Shanghai STAR (科创板)
    def test_shanghai_688(self):
        assert to_tushare_code("688001") == "688001.SH"

    # Shanghai B-share
    def test_shanghai_900(self):
        assert to_tushare_code("900901") == "900901.SH"

    # Shenzhen main board
    def test_shenzhen_000(self):
        assert to_tushare_code("000890") == "000890.SZ"

    def test_shenzhen_001(self):
        assert to_tushare_code("001234") == "001234.SZ"

    # Shenzhen SME (中小板)
    def test_shenzhen_002(self):
        assert to_tushare_code("002709") == "002709.SZ"

    # Shenzhen ChiNext (创业板)
    def test_shenzhen_300(self):
        assert to_tushare_code("300342") == "300342.SZ"

    # Shenzhen B-share
    def test_shenzhen_200(self):
        assert to_tushare_code("200002") == "200002.SZ"

    # Shanghai ETF
    def test_shanghai_etf_510(self):
        assert to_tushare_code("510050") == "510050.SH"

    def test_shanghai_etf_512(self):
        assert to_tushare_code("512100") == "512100.SH"

    def test_shanghai_etf_515(self):
        assert to_tushare_code("515030") == "515030.SH"

    # Shenzhen ETF
    def test_shenzhen_etf_159(self):
        assert to_tushare_code("159915") == "159915.SZ"

    # Pass-through for already-suffixed codes
    def test_already_suffixed_sh(self):
        assert to_tushare_code("600537.SH") == "600537.SH"

    def test_already_suffixed_sz(self):
        assert to_tushare_code("000890.SZ") == "000890.SZ"

    # Invalid / IPO subscription codes
    def test_ipo_730_raises(self):
        with pytest.raises(ValueError, match="Cannot determine market"):
            to_tushare_code("730123")

    def test_unknown_prefix_raises(self):
        with pytest.raises(ValueError, match="Cannot determine market"):
            to_tushare_code("500123")

    def test_4digit_short_code_raises(self):
        with pytest.raises(ValueError, match="must be a 6-digit code"):
            to_tushare_code("60")
