package cron

import (
	"context"

	robfigcron "github.com/robfig/cron/v3"

	"accounting/internal/adapter/config"
	"accounting/internal/dto"
	"accounting/internal/port"
	"accounting/internal/service"
)

// Scheduler manages periodic banking synchronization
type Scheduler struct {
	cron        *robfigcron.Cron
	syncService *service.StatementSyncService
	logger      port.Logger
	cfg         config.SyncJobConfig
}

func NewScheduler(
	cfg config.SyncJobConfig,
	syncService *service.StatementSyncService,
	logger port.Logger,
) *Scheduler {
	return &Scheduler{
		cron:        robfigcron.New(),
		syncService: syncService,
		logger:      logger,
		cfg:         cfg,
	}
}

func (s *Scheduler) Start() error {
	if !s.cfg.Enabled {
		s.logger.IPrintf(0, "Sync job cron is disabled in config.")
		return nil
	}

	cronExpr := s.cfg.CronExpr
	if cronExpr == "" {
		cronExpr = "0 */6 * * *"
	}

	_, err := s.cron.AddFunc(cronExpr, func() {
		s.logger.IPrintf(0, "Executing scheduled banking statement sync (cron: %s)...", cronExpr)
		resp := s.syncService.Run(context.Background(), &dto.SyncStatementsRequest{
			DaysBack: s.cfg.DaysBack,
		})
		s.logger.IPrintf(0, "Scheduled sync finished: %d imported, %d ignored", resp.TotalImported, resp.TotalIgnored)
	})

	if err != nil {
		return err
	}

	s.cron.Start()
	s.logger.IPrintf(0, "Sync scheduler started with cron expression: %s", cronExpr)
	return nil
}

func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
}

