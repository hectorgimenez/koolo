package bot

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/event"
)

const (
	NotStarted SupervisorStatus = "Not Started"
	Starting   SupervisorStatus = "Starting"
	InGame     SupervisorStatus = "In game"
	Paused     SupervisorStatus = "Paused"
	Crashed    SupervisorStatus = "Crashed"
)

type SupervisorStatus string

type StatsHandler struct {
	stats  *Stats
	name   string
	logger *slog.Logger
}

func NewStatsHandler(name string, logger *slog.Logger) *StatsHandler {
	return &StatsHandler{
		name:   name,
		logger: logger,
		stats: &Stats{
			SupervisorStatus: Starting,
			StartedAt:        time.Now(),
		},
	}
}

func (h *StatsHandler) Handle(_ context.Context, e event.Event) error {
	// Only handle events from the supervisor
	if !strings.EqualFold(e.Supervisor(), h.name) {
		return nil
	}

	switch evt := e.(type) {
	case event.GameCreatedEvent:
		h.stats.Games = append(h.stats.Games, GameStats{
			StartedAt: evt.OccurredAt(),
		})
		h.stats.SupervisorStatus = InGame

	case event.GameFinishedEvent:
		if len(h.stats.Games) > 0 {
			h.stats.Games[len(h.stats.Games)-1].FinishedAt = evt.OccurredAt()
			h.stats.Games[len(h.stats.Games)-1].Reason = evt.Reason
		}

	case event.RunStartedEvent:
		if len(h.stats.Games) > 0 {
			h.stats.Games[len(h.stats.Games)-1].Runs = append(h.stats.Games[len(h.stats.Games)-1].Runs, RunStats{
				Name:      evt.RunName,
				StartedAt: evt.OccurredAt(),
			})
		}

	case event.RunFinishedEvent:
		if len(h.stats.Games) > 0 && len(h.stats.Games[len(h.stats.Games)-1].Runs) > 0 {
			lastRun := &h.stats.Games[len(h.stats.Games)-1].Runs[len(h.stats.Games[len(h.stats.Games)-1].Runs)-1]
			lastRun.FinishedAt = evt.OccurredAt()
			lastRun.Reason = evt.Reason
		}

	case event.GamePausedEvent:
		if evt.Paused {
			h.stats.SupervisorStatus = Paused
		} else {
			h.stats.SupervisorStatus = InGame
		}

	case event.ItemStashedEvent:
		h.stats.Drops = append(h.stats.Drops, evt.Item)

	case event.UsedPotionEvent:
		if len(h.stats.Games) > 0 && len(h.stats.Games[len(h.stats.Games)-1].Runs) > 0 {
			lastRun := &h.stats.Games[len(h.stats.Games)-1].Runs[len(h.stats.Games[len(h.stats.Games)-1].Runs)-1]
			lastRun.UsedPotions = append(lastRun.UsedPotions, evt)
		}
	}

	return nil
}

func (h *StatsHandler) Stats() Stats {
	return *h.stats
}

type Stats struct {
	StartedAt           time.Time
	SupervisorStatus    SupervisorStatus
	Details             string
	Drops               []data.Drop
	Games               []GameStats
	IsCompanionFollower bool
}

type GameStats struct {
	StartedAt  time.Time
	FinishedAt time.Time
	Reason     event.FinishReason
	Runs       []RunStats
}

type RunStats struct {
	Name        string
	Reason      event.FinishReason
	StartedAt   time.Time
	Items       []data.Item
	FinishedAt  time.Time
	UsedPotions []event.UsedPotionEvent
}

func (s Stats) TotalGames() int {
	return len(s.Games)
}

func (s Stats) TotalDeaths() int {
	return s.totalRunsByReason(event.FinishedDied)
}

func (s Stats) TotalChickens() int {
	return s.totalRunsByReason(event.FinishedChicken) + s.totalRunsByReason(event.FinishedMercChicken)
}

func (s Stats) TotalErrors() int {
	return s.totalRunsByReason(event.FinishedError)
}

func (s Stats) totalRunsByReason(reason event.FinishReason) int {
	total := 0
	for _, g := range s.Games {
		for _, r := range g.Runs {
			if r.Reason == reason {
				total++
			}
		}
	}

	return total
}
