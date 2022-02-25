package stats

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"strings"
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

	Status.TotalGames++
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

func PickupItem(item game.Item) {
	if item.IsPotion() || strings.EqualFold(item.Name, "Gold") {
		return
	}

	Status.RunStats[Status.CurrentRun].ItemsFound = append(Status.RunStats[Status.CurrentRun].ItemsFound, item)
}

type GameStatus struct {
	ApplicationStartedAt time.Time
	TotalGames           int
	RunStats             map[string]*RunStats

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
	ItemsFound             []game.Item
	HealingPotionsUsed     int
	ManaPotionsUsed        int
	RejuvPotionsUsed       int
	MercHealingPotionsUsed int
	MercRejuvPotionsUsed   int
}
