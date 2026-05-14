package scheduler_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stock/pkg/stockd/db"
	"stock/pkg/stockd/models"
	"stock/pkg/stockd/services/scheduler"
)

func openDB(t *testing.T) *gorm.DB {
	gdb, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(gdb))
	return gdb
}

func TestTrigger_RecordsJobRun(t *testing.T) {
	gdb := openDB(t)
	s := scheduler.New(gdb)
	var calls int32
	s.RegisterFunc("test", func(ctx context.Context) error { atomic.AddInt32(&calls, 1); return nil })

	require.NoError(t, s.Trigger(context.Background(), "test"))
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))

	var rows []models.JobRun
	require.NoError(t, gdb.Find(&rows).Error)
	require.Len(t, rows, 1)
	assert.Equal(t, "success", rows[0].Status)
	require.NotNil(t, rows[0].FinishedAt)
}

func TestTrigger_DeDupesConcurrent(t *testing.T) {
	gdb := openDB(t)
	s := scheduler.New(gdb)
	started := make(chan struct{})
	release := make(chan struct{})
	var calls int32
	s.RegisterFunc("slow", func(ctx context.Context) error {
		atomic.AddInt32(&calls, 1)
		started <- struct{}{}
		<-release
		return nil
	})

	go func() { _ = s.Trigger(context.Background(), "slow") }()
	<-started
	go func() { _ = s.Trigger(context.Background(), "slow") }()
	time.Sleep(50 * time.Millisecond)
	close(release)

	assert.Equal(t, int32(1), atomic.LoadInt32(&calls), "singleflight collapses the second call")
}
