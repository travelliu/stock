// Package utils converts plain A-share codes (e.g. "600537") into
// Tushare-suffixed codes ("600537.SH" / "000890.SZ").
package utils

import (
	"strings"
)

var shPrefixes = map[string]struct{}{
	"600": {}, "601": {}, "603": {}, "605": {}, "688": {},
	"900": {},
	"510": {}, "511": {}, "512": {}, "513": {}, "515": {},
}

var szPrefixes = map[string]struct{}{
	"000": {}, "001": {}, "002": {}, "300": {},
	"200": {},
	"159": {},
}

// ToTushareCode converts a plain 6-digit A-share code to the Tushare suffix
// form. Already-suffixed inputs (containing ".") pass through unchanged.
func ToTushareCode(code string) string {
	if strings.Contains(code, ".") {
		return code
	}
	if len(code) < 6 {
		return code
	}
	prefix := code[:3]
	if _, ok := shPrefixes[prefix]; ok {
		return code + ".SH"
	}
	if _, ok := szPrefixes[prefix]; ok {
		return code + ".SZ"
	}
	return code
}
