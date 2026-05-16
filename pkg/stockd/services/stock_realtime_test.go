package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsTradingHours(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")

	cases := []struct {
		name string
		t    time.Time
		want bool
	}{
		{
			name: "weekday at 9:15 (open)",
			t:    time.Date(2026, 5, 12, 9, 15, 0, 0, loc), // Monday
			want: true,
		},
		{
			name: "weekday at 15:00 (close)",
			t:    time.Date(2026, 5, 12, 15, 0, 0, 0, loc),
			want: true,
		},
		{
			name: "weekday at 9:14 (before open)",
			t:    time.Date(2026, 5, 12, 9, 14, 0, 0, loc),
			want: false,
		},
		{
			name: "weekday at 15:01 (after close)",
			t:    time.Date(2026, 5, 12, 15, 1, 0, 0, loc),
			want: false,
		},
		{
			name: "Saturday",
			t:    time.Date(2026, 5, 16, 10, 0, 0, 0, loc),
			want: false,
		},
		{
			name: "Sunday",
			t:    time.Date(2026, 5, 17, 10, 0, 0, 0, loc),
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isTradingHours(tc.t))
		})
	}
}
