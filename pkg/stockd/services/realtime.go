package services

import (
	"time"
)

// isTradingHours reports whether t falls within 09:15–15:00 on a weekday in Asia/Shanghai.
func isTradingHours(t time.Time) bool {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := t.In(loc)
	wd := now.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return false
	}
	h, m, _ := now.Clock()
	total := h*60 + m
	return total >= 9*60+15 && total <= 15*60
}
