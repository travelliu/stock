package services_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/models"
	"stock/pkg/stockd/services"
)

func TestListPredictionsPage(t *testing.T) {
	gdb := openDB(t)
	for i := 1; i <= 25; i++ {
		p := models.AnalysisPrediction{
			TsCode:         "X.SH",
			TradeDate:      fmt.Sprintf("2025%04d", i),
			OpenPrice:      float64(i),
			SampleCounts:   json.RawMessage(`{}`),
			WindowMeans:    json.RawMessage(`[]`),
			CompositeMeans: json.RawMessage(`{}`),
		}
		require.NoError(t, gdb.Create(&p).Error)
	}
	svc := services.New(gdb)
	page, err := svc.ListPredictionsPage(context.Background(), "X.SH", "", "", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(25), page.Total)
	assert.Len(t, page.Items, 20)
	page2, _ := svc.ListPredictionsPage(context.Background(), "X.SH", "", "", 2, 20)
	assert.Len(t, page2.Items, 5)
}
