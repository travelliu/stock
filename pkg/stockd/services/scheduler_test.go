package services_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/models"
)

func TestTrigger_RecordsJobRun(t *testing.T) {
	gdb := openDB(t)
	s := newService(t)
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
	s := newService(t)
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
