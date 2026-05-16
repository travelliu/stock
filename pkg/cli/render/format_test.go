package render_test

import (
	"stock/pkg/cli/render"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisplayWidth(t *testing.T) {
	assert.Equal(t, 5, render.DisplayWidth("hello"))
	assert.Equal(t, 4, render.DisplayWidth("开盘"))
	assert.Equal(t, 3, render.DisplayWidth("开A"))
}

func TestFormatTable_Shape(t *testing.T) {
	out := render.FormatTable([]string{"时段", "数值"}, [][]string{{"历史", "1.23"}, {"近1月", "0.45"}})
	assert.Contains(t, out, "时段")
	assert.Contains(t, out, "历史")
	assert.Contains(t, out, "+")
	lines := 0
	for _, c := range out {
		if c == '\n' {
			lines++
		}
	}
	assert.Equal(t, 5, lines, "6 lines means 5 newlines")
}
