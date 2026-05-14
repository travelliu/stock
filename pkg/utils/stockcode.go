// Package utils converts plain A-share codes (e.g. "600537") into
// Tushare-suffixed codes ("600537.SH" / "000890.SZ").
package utils

import (
	"fmt"
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
func ToTushareCode(code string) (string, error) {
	if strings.Contains(code, ".") {
		return code, nil
	}
	if len(code) < 6 {
		return "", fmt.Errorf("invalid stock code %q: must be 6 digits", code)
	}
	prefix := code[:3]
	if _, ok := shPrefixes[prefix]; ok {
		return code + ".SH", nil
	}
	if _, ok := szPrefixes[prefix]; ok {
		return code + ".SZ", nil
	}
	return "", fmt.Errorf("cannot determine market for stock code %q (prefix %q is not a known SH/SZ prefix)", code, prefix)
}
