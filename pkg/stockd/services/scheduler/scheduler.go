// Package scheduler hosts cron jobs guarded by singleflight, records job_runs.
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"stock/pkg/stockd/models"
)

// JobFunc is the signature for registered jobs.
type JobFunc func(ctx context.Context) error

type Service struct {
	db   *gorm.DB
	cron *cron.Cron
	mu   sync.Mutex
	jobs map[string]JobFunc
	sf   singleflight.Group
}

func New(db *gorm.DB) *Service {
	return &Service{
		db:   db,
		cron: cron.New(cron.WithLocation(time.Local)),
		jobs: map[string]JobFunc{},
	}
}

// RegisterFunc registers a JobFunc without a cron schedule (manual-trigger only).
func (s *Service) RegisterFunc(name string, fn JobFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[name] = fn
}

// RegisterCron registers a JobFunc on a cron schedule.
func (s *Service) RegisterCron(name, expr string, fn JobFunc) error {
	s.RegisterFunc(name, fn)
	_, err := s.cron.AddFunc(expr, func() { _ = s.Trigger(context.Background(), name) })
	return err
}

func (s *Service) Start() { s.cron.Start() }
func (s *Service) Stop()  { _ = s.cron.Stop() }

// Trigger runs the named job through singleflight + records a JobRun row.
// Returns an error if the job is unknown OR returned an error.
func (s *Service) Trigger(ctx context.Context, name string) error {
	s.mu.Lock()
	fn, ok := s.jobs[name]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("unknown job %q", name)
	}
	_, err, _ := s.sf.Do(name, func() (any, error) {
		run := &models.JobRun{JobName: name, StartedAt: time.Now(), Status: "running"}
		if err := s.db.Create(run).Error; err != nil {
			return nil, err
		}
		jobErr := fn(ctx)
		end := time.Now()
		run.FinishedAt = &end
		if jobErr != nil {
			run.Status = "error"
			run.Message = jobErr.Error()
		} else {
			run.Status = "success"
		}
		_ = s.db.Save(run).Error
		return nil, jobErr
	})
	return err
}

// LastRun returns the most recent JobRun row for a named job.
func (s *Service) LastRun(ctx context.Context, name string) (*models.JobRun, error) {
	var row models.JobRun
	err := s.db.WithContext(ctx).Where("job_name = ?", name).
		Order("started_at DESC").Limit(1).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}
