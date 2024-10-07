package bot

import (
	"log/slog"
	"time"

	"github.com/hectorgimenez/koolo/internal/config"
)

type Scheduler struct {
	manager *SupervisorManager
	logger  *slog.Logger
	stop    chan struct{}
}

func NewScheduler(manager *SupervisorManager, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		manager: manager,
		logger:  logger,
		stop:    make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	s.logger.Info("Scheduler started")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkSchedules()
		case <-s.stop:
			s.logger.Info("Scheduler stopped")
			return
		}
	}
}

func (s *Scheduler) Stop() {
	close(s.stop)
}

func (s *Scheduler) checkSchedules() {
	now := time.Now()
	currentDay := int(now.Weekday())

	for supervisorName, cfg := range config.Characters {
		if !cfg.Scheduler.Enabled {
			continue
		}

		for _, day := range cfg.Scheduler.Days {
			if day.DayOfWeek != currentDay {
				continue
			}

			actionTaken := false

			// Check if any time range is active
			for _, timeRange := range day.TimeRanges {
				start := time.Date(now.Year(), now.Month(), now.Day(), timeRange.Start.Hour(), timeRange.Start.Minute(), 0, 0, now.Location())
				end := time.Date(now.Year(), now.Month(), now.Day(), timeRange.End.Hour(), timeRange.End.Minute(), 0, 0, now.Location())

				if now.After(start) && now.Before(end) && s.supervisorNotStarted(supervisorName) {
					s.logger.Info("Starting supervisor based on schedule. Time range: "+start.Format("15:04")+" - "+end.Format("15:04"), "supervisor", supervisorName)
					go s.startSupervisor(supervisorName)
					actionTaken = true
					break
				} else if now.After(end) || now == end || now.Before(start) && !s.supervisorNotStarted(supervisorName) {
					s.logger.Info("Stopping supervisor based on schedule. Time range: "+start.Format("15:04")+" - "+end.Format("15:04"), "supervisor", supervisorName)
					s.stopSupervisor(supervisorName)
					actionTaken = true
					break
				}
			}

			// To prevent unnecessary checks since it can be only one day.
			// Also avoid starting/stoping a supervisor in the same minute ... this is rediculous
			if actionTaken {
				break
			}
		}
	}
}

func (s *Scheduler) supervisorNotStarted(name string) bool {
	stats := s.manager.GetSupervisorStats(name)
	return stats.SupervisorStatus == NotStarted || stats.SupervisorStatus == Crashed || stats.SupervisorStatus == ""
}

func (s *Scheduler) startSupervisor(name string) {
	if s.supervisorNotStarted(name) {
		err := s.manager.Start(name, false)
		if err != nil {
			s.logger.Error("Failed to start supervisor", "supervisor", name, "error", err)
		}
	}
}

func (s *Scheduler) stopSupervisor(name string) {
	if !s.supervisorNotStarted(name) {
		s.manager.Stop(name)
	}
}
