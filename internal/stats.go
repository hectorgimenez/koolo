package koolo

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
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
		h.stats.Games[len(h.stats.Games)-1].FinishedAt = evt.OccurredAt()
		h.stats.Games[len(h.stats.Games)-1].Reason = evt.Reason
		h.stats.Games[len(h.stats.Games)-1].Runs[len(h.stats.Games[len(h.stats.Games)-1].Runs)-1].FinishedAt = evt.OccurredAt()
		h.stats.Games[len(h.stats.Games)-1].Runs[len(h.stats.Games[len(h.stats.Games)-1].Runs)-1].Reason = evt.Reason
	case event.RunStartedEvent:
		if len(h.stats.Games) == 0 {
			h.stats.Games = append(h.stats.Games, GameStats{
				StartedAt: evt.OccurredAt(),
			})
			h.stats.SupervisorStatus = InGame
		}
		h.stats.Games[len(h.stats.Games)-1].Runs = append(h.stats.Games[len(h.stats.Games)-1].Runs, RunStats{
			Name:      evt.RunName,
			StartedAt: evt.OccurredAt(),
		})
	case event.GamePausedEvent:
		if evt.Paused {
			h.stats.SupervisorStatus = Paused
		} else {
			h.stats.SupervisorStatus = InGame
		}
	case event.RunFinishedEvent:
		h.stats.Games[len(h.stats.Games)-1].Runs[len(h.stats.Games[len(h.stats.Games)-1].Runs)-1].FinishedAt = evt.OccurredAt()
		h.stats.Games[len(h.stats.Games)-1].Runs[len(h.stats.Games[len(h.stats.Games)-1].Runs)-1].Reason = evt.Reason
	case event.ItemStashedEvent:
		// The hell is this Hector o.O
		//h.stats.Games[len(h.stats.Games)-1].Runs[len(h.stats.Games[len(h.stats.Games)-1].Runs)-1].Items = append(h.stats.Games[len(h.stats.Games)-1].Runs[len(h.stats.Games[len(h.stats.Games)-1].Runs)-1].Items, evt.Item)

		// Ain't this much easier?
		h.stats.Drops = append(h.stats.Drops, evt.Item)
	case event.UsedPotionEvent:
		h.stats.Games[len(h.stats.Games)-1].Runs[len(h.stats.Games[len(h.stats.Games)-1].Runs)-1].UsedPotions = append(h.stats.Games[len(h.stats.Games)-1].Runs[len(h.stats.Games[len(h.stats.Games)-1].Runs)-1].UsedPotions, evt)
	}

	return nil
}

func (h *StatsHandler) updateGameStatsFile() {
	if _, err := os.Stat("stats"); os.IsNotExist(err) {
		err = os.MkdirAll("stats", os.ModePerm)
		if err != nil {
			h.logger.Error("Error creating stats directory", slog.Any("error", err))
			return
		}
	}

	fileName := fmt.Sprintf("stats/stats_%s_%s.txt", h.name, h.stats.StartedAt.Format("2006-02-01-15_04_05"))
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		h.logger.Error("Error writing game stats", slog.Any("error", err))
		return
	}
	w := bufio.NewWriter(f)

	//for _, game := range h.stats.Games {
	//	var items = ""
	//	for _, item := range rs.ItemsFound {
	//		items += fmt.Sprintf("%s [%d]\n", item.Name, item.Quality)
	//	}
	//	avgRunTime := rs.RunningTime.Seconds() / float64(rs.Errors+rs.Kills+rs.Deaths+rs.Chickens+rs.MerChicken)
	//	statsRun := fmt.Sprintf("Stats for: %s\n"+
	//		"    Run time: %0.2fs (Total) %0.2fs (Average)\n"+
	//		"    Kills: %d\n"+
	//		"    Deaths: %d\n"+
	//		"    Chickens: %d\n"+
	//		"    Merc Chickens: %d\n"+
	//		"    Errors: %d\n"+
	//		"    Used HP Potions: %d\n"+
	//		"    Used MP Potions: %d\n"+
	//		"    Used Rejuv Potions: %d\n"+
	//		"    Used Merc HP Potions: %d\n"+
	//		"    Used Merc Rejuv Potions: %d\n"+
	//		"    Items: \n"+
	//		"    %s",
	//		runName,
	//		rs.RunningTime.Seconds(), avgRunTime,
	//		rs.Kills,
	//		rs.Deaths,
	//		rs.Chickens,
	//		rs.MerChicken,
	//		rs.Errors,
	//		rs.HealingPotionsUsed,
	//		rs.ManaPotionsUsed,
	//		rs.RejuvPotionsUsed,
	//		rs.MercHealingPotionsUsed,
	//		rs.MercRejuvPotionsUsed,
	//		items,
	//	)
	//	_, err = w.WriteString(statsRun + "\n")
	//	if err != nil {
	//		s.logger.Error("Error writing stats file", slog.Any("error", err))
	//	}
	//}

	w.Flush()
	f.Close()
}

func (h *StatsHandler) Stats() Stats {
	return *h.stats
}

type Stats struct {
	StartedAt        time.Time
	SupervisorStatus SupervisorStatus
	Details          string
	Drops            []data.Drop
	Games            []GameStats
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
