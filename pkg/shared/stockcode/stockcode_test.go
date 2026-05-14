package stockcode_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/shared/stockcode"
)

func TestToTushareCode(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"sh-600", "600537", "600537.SH", false},
		{"sh-601", "601398", "601398.SH", false},
		{"sh-603", "603778", "603778.SH", false},
		{"sh-605", "605588", "605588.SH", false},
		{"sh-688-star", "688001", "688001.SH", false},
		{"sh-900-b", "900901", "900901.SH", false},
		{"sh-510-etf", "510300", "510300.SH", false},
		{"sh-515-etf", "515170", "515170.SH", false},
		{"sz-000", "000001", "000001.SZ", false},
		{"sz-001", "001872", "001872.SZ", false},
		{"sz-002", "002594", "002594.SZ", false},
		{"sz-300-gem", "300750", "300750.SZ", false},
		{"sz-200-b", "200012", "200012.SZ", false},
		{"sz-159-etf", "159915", "159915.SZ", false},
		{"passthrough-suffixed", "600537.SH", "600537.SH", false},
		{"passthrough-sz", "000890.SZ", "000890.SZ", false},
		{"too-short", "12345", "", true},
		{"ipo-subscription-730", "730001", "", true},
		{"ipo-subscription-732", "732001", "", true},
		{"unknown-prefix-400", "400001", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := stockcode.ToTushareCode(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
