package cron

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"

	"billing/internal/dto"
	"billing/internal/port"
	"billing/internal/service"
)

// Scheduler represents the cron driver adapter for scheduling periodic tasks.
type Scheduler struct {
	cron          *cron.Cron
	autoBillSvc   *service.InvoiceAutoBill
	logger        port.Logger
	daysInAdvance int
	schedules     []string
	isRunning     atomic.Bool
	entryIDs      []cron.EntryID
}

// NewScheduler creates a new instance of Scheduler with timezone and overlap protection.
func NewScheduler(autoBillSvc *service.InvoiceAutoBill, logger port.Logger,
	daysInAdvance int, schedules []string, timezone string) (*Scheduler, error) {

	loc := time.Local
	if timezone != "" {
		loadedLoc, err := time.LoadLocation(timezone)
		if err != nil {
			logger.IPrintf(0, "CronScheduler: Failed to load timezone %s (%v), using Local", timezone, err)
		} else {
			loc = loadedLoc
		}
	}

	c := cron.New(
		cron.WithLocation(loc),
	)

	return &Scheduler{
		cron:          c,
		autoBillSvc:   autoBillSvc,
		logger:        logger,
		daysInAdvance: daysInAdvance,
		schedules:     schedules,
	}, nil
}

// Start registers all configured schedules and starts the cron scheduler.
func (s *Scheduler) Start() error {
	for _, schedule := range s.schedules {
		sched := schedule
		entryID, err := s.cron.AddFunc(sched, s.ExecuteJob)
		if err != nil {
			s.logger.IPrintf(0, "CronScheduler: Failed to register schedule %q: %v", sched, err)
			return fmt.Errorf("failed to register cron schedule %q: %w", sched, err)
		}
		s.entryIDs = append(s.entryIDs, entryID)
		s.logger.IPrintf(0, "CronScheduler: Registered auto-bill schedule %q (Entry ID: %d)", sched, entryID)
	}

	s.cron.Start()
	s.logger.IPrintf(0, "CronScheduler: Started successfully with %d schedule(s)", len(s.entryIDs))
	return nil
}

// Stop stops the cron scheduler.
func (s *Scheduler) Stop() error {
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.IPrintf(0, "CronScheduler: Stopped successfully")
	return nil
}

// ExecuteJob executes the auto-bill service while preventing overlapping runs.
func (s *Scheduler) ExecuteJob() {
	if !s.isRunning.CompareAndSwap(false, true) {
		s.logger.IPrintf(0, "CronScheduler: Execution skipped. Previous auto-bill job is still in progress (overlap detected)")
		return
	}
	defer s.isRunning.Store(false)

	s.logger.IPrintf(1, "CronScheduler: Starting scheduled auto-bill execution...")
	req := &dto.InvoiceAutoBillRequest{
		DaysInAdvance: s.daysInAdvance,
	}

	res := s.autoBillSvc.Run(req)
	if resp, ok := res.(*dto.InvoiceAutoBillResponse); ok {
		s.logger.IPrintf(1, "CronScheduler: Execution finished (Found: %d, Sent: %d, Errors: %d)",
			resp.TotalFound, resp.TotalSent, resp.TotalErrors)
	} else {
		s.logger.IPrintf(1, "CronScheduler: Execution finished with status code %d", res.GetStatusCode())
	}
}

// IsJobRunning returns whether an auto-bill execution is currently active.
func (s *Scheduler) IsJobRunning() bool {
	return s.isRunning.Load()
}
