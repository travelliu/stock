package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"stock/pkg/utils"
)

func TestToTushareCode(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"sh-600", "600537", "600537.SH"},
		{"sh-601", "601398", "601398.SH"},
		{"sh-603", "603778", "603778.SH"},
		{"sh-605", "605588", "605588.SH"},
		{"sh-688-star", "688001", "688001.SH"},
		{"sh-900-b", "900901", "900901.SH"},
		{"sh-510-etf", "510300", "510300.SH"},
		{"sh-515-etf", "515170", "515170.SH"},
		{"sz-000", "000001", "000001.SZ"},
		{"sz-001", "001872", "001872.SZ"},
		{"sz-002", "002594", "002594.SZ"},
		{"sz-300-gem", "300750", "300750.SZ"},
		{"sz-200-b", "200012", "200012.SZ"},
		{"sz-159-etf", "159915", "159915.SZ"},
		{"passthrough-suffixed", "600537.SH", "600537.SH"},
		{"passthrough-sz", "000890.SZ", "000890.SZ"},
		{"too-short", "12345", "12345"},
		{"ipo-subscription-730", "730001", "730001"},
		{"ipo-subscription-732", "732001", "732001"},
		{"unknown-prefix-400", "400001", "400001"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := utils.ToTushareCode(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
