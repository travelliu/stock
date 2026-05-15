package analysis_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stock/pkg/stockd/services/analysis"
)

func TestA1(t *testing.T) {
	fmt.Println(os.Getwd())

	gdb, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?cache=shared", "/root/code/github/travelliu/stock/data/stock.db")),
		&gorm.Config{})
	require.NoError(t, err)
	s := analysis.New(gdb, nil)
	s.Run(context.Background(), analysis.Input{UserID: 1, TsCode: "300476"})
}
