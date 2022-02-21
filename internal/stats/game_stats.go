package stats

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"time"
)

var Status = GameStatus{}

const (
	EventKill        Event = "kill"
	EventDeath       Event = "death"
	EventChicken     Event = "chicken"
	EventMercChicken Event = "merc chicken"
	EventError       Event = "error"
)

func StartRun(runName string) {
	if Status.RunStats == nil {
		Status.RunStats = map[string]*RunStats{}
	}

	rs, found := Status.RunStats[runName]
	if !found {
		rs = &RunStats{}
		Status.RunStats[runName] = rs
	}

	Status.CurrentRun = runName
	Status.CurrentRunStart = time.Now()
}

func FinishCurrentRun(event Event) {
	rs := Status.RunStats[Status.CurrentRun]
	switch event {
	case EventKill:
		rs.Kills++
	case EventDeath:
		rs.Deaths++
	case EventChicken:
		rs.Chickens++
	case EventMercChicken:
		rs.MerChicken++
	case EventError:
		rs.Errors++
	}

	runTime := time.Since(Status.CurrentRunStart)
	rs.TotalRunsTime += runTime
	Status.TotalTime += runTime
}

func UsedPotion(potionType game.PotionType, onMerc bool) {
	switch potionType {
	case game.HealingPotion:
		if onMerc {
			Status.RunStats[Status.CurrentRun].MercHealingPotionsUsed++
		} else {
			Status.RunStats[Status.CurrentRun].HealingPotionsUsed++
		}
	case game.ManaPotion:
		Status.RunStats[Status.CurrentRun].ManaPotionsUsed++
	case game.RejuvenationPotion:
		if onMerc {
			Status.RunStats[Status.CurrentRun].MercRejuvPotionsUsed++
		} else {
			Status.RunStats[Status.CurrentRun].RejuvPotionsUsed++
		}
	}
}

type GameStatus struct {
	TotalTime time.Duration
	RunStats  map[string]*RunStats

	CurrentRun      string
	CurrentRunStart time.Time
}

type Event string
type RunStats struct {
	TotalRunsTime          time.Duration
	Kills                  int
	Deaths                 int
	Chickens               int
	MerChicken             int
	Errors                 int
	ItemsFound             int
	HealingPotionsUsed     int
	ManaPotionsUsed        int
	RejuvPotionsUsed       int
	MercHealingPotionsUsed int
	MercRejuvPotionsUsed   int
}
