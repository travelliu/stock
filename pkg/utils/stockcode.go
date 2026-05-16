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
	if strings.HasPrefix(code, "sz") {
		return strings.TrimPrefix(code, "sz") + ".SZ"
	}
	if strings.HasPrefix(code, "sh") {
		return strings.TrimPrefix(code, "sh") + ".sh"
	}
	codeType := GetStockType(code)
	if codeType == "" {
		return code
	}
	return code + "." + codeType
}

func GetStockType(code string) string {
	prefix := code[:3]
	if _, ok := shPrefixes[prefix]; ok {
		return "SH"
	}
	if _, ok := szPrefixes[prefix]; ok {
		return "SZ"
	}
	return ""
}

func ToTencentCode(code string) string {
	if strings.Contains(code, ".") {
		parts := strings.SplitN(code, ".", 2)
		if len(parts) != 2 {
			return code
		}
		return strings.ToLower(parts[1]) + parts[0]
	}
	codeType := GetStockType(code)
	if codeType == "" {
		return code
	}
	return strings.ToLower(codeType) + code
}
