// 2026/5/14 Bin Liu <bin.liu@enmotech.com>

package utils

import (
	"strings"
)

func TrimTsCode(tsCode string) string {
	s := strings.Split(tsCode, ".")
	if len(s) != 2 {
		return tsCode
	}
	return s[0]
}
