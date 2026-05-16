package tencent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixture has exactly 50 fields (indices 0-49): 量比 is at index 49
const fixtureResponse = `v_sh600519="1~贵州茅台~600519~1780.00~1775.00~1778.00~12345~200~100~1779.00~100~1779.50~200~1779.80~300~1780.00~100~1780.00~50~1780.10~100~1780.20~200~1780.30~300~1780.40~100~1780.50~50~recent~14:58:30~5.00~0.28~1785.00~1760.00~stuff~12345~5000.00~1.23~28.5~field40~1800.00~1600.00~1.50~2000.00~2500.00~5.20~1815.80~1738.60~1.85";`

func TestFetchQuotes_ParsesFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fixtureResponse))
	}))
	defer srv.Close()
	
	c := NewClient(WithBaseURL(srv.URL + "/q="))
	quotes, err := c.FetchQuotes(context.Background(), []string{"600519"})
	require.NoError(t, err)
	require.Len(t, quotes, 1)
	
	q := quotes[0]
	assert.Equal(t, "600519", q.TsCode)
	assert.Empty(t, q.Name) // GBK name skipped; fillNames populates from DB cache
	assert.InDelta(t, 1780.00, q.Price, 0.001)
	assert.InDelta(t, 1775.00, q.PrevClose, 0.001)
	assert.InDelta(t, 1778.00, q.Open, 0.001)
	assert.InDelta(t, 12345.0, q.Vol, 0.001)
	assert.InDelta(t, 200.0, q.OuterVol, 0.001)
	assert.InDelta(t, 100.0, q.InnerVol, 0.001)
	assert.InDelta(t, 1785.00, q.High, 0.001)
	assert.InDelta(t, 1760.00, q.Low, 0.001)
	assert.InDelta(t, 12345.0, q.TotalVol, 0.001)
	assert.InDelta(t, 5000.00, q.Amount, 0.001)
	assert.InDelta(t, 1.23, q.TurnoverRate, 0.001)
	assert.InDelta(t, 28.5, q.PE, 0.001)
	assert.InDelta(t, 1800.00, q.High52w, 0.001)
	assert.InDelta(t, 1600.00, q.Low52w, 0.001)
	assert.InDelta(t, 1.50, q.Amplitude, 0.001)
	assert.InDelta(t, 2000.00, q.CircMarketCap, 0.001)
	assert.InDelta(t, 2500.00, q.TotalMarketCap, 0.001)
	assert.InDelta(t, 5.20, q.PB, 0.001)
	assert.InDelta(t, 5.00, q.Change, 0.001)
	assert.InDelta(t, 0.28, q.ChangePct, 0.001)
	assert.InDelta(t, 1815.80, q.LimitUp, 0.001)
	assert.InDelta(t, 1738.60, q.LimitDown, 0.001)
	assert.InDelta(t, 1.85, q.VolRatio, 0.001)
	assert.Equal(t, "14:58:30", q.QuoteTime)
}

func TestTsToCodes(t *testing.T) {
	got := tsToCodes([]string{"600519.SH", "000858"})
	assert.Equal(t, []string{"sh600519", "sz000858"}, got)
}

func TestTencentToTs(t *testing.T) {
	cases := []struct{ in, want string }{
		{"sh600519", "600519"},
		{"sz000858", "000858"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, tencentToTs(tc.in))
	}
}

func TestFetchQuotes_EmptyInput(t *testing.T) {
	c := NewClient()
	quotes, err := c.FetchQuotes(context.Background(), []string{})
	require.NoError(t, err)
	assert.Empty(t, quotes)
}
