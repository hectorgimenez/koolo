package stat

import (
	"fmt"
	"image"
	"strings"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/event"
)

var Status = GameStatus{}

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

func FinishCurrentRun(evt event.Event) {
	rs := Status.RunStats[Status.CurrentRun]
	switch evt {
	case event.Kill:
		rs.Kills++
	case event.Death:
		rs.Deaths++
	case event.Chicken:
		rs.Chickens++
	case event.MercChicken:
		rs.MerChicken++
	case event.Error:
		rs.Errors++
	}

	runTime := time.Since(Status.CurrentRunStart)
	rs.TotalRunsTime += runTime
}

func UsedPotion(potionType data.PotionType, onMerc bool) {
	switch potionType {
	case data.HealingPotion:
		if onMerc {
			Status.RunStats[Status.CurrentRun].MercHealingPotionsUsed++
		} else {
			Status.RunStats[Status.CurrentRun].HealingPotionsUsed++
		}
	case data.ManaPotion:
		Status.RunStats[Status.CurrentRun].ManaPotionsUsed++
	case data.RejuvenationPotion:
		if onMerc {
			Status.RunStats[Status.CurrentRun].MercRejuvPotionsUsed++
		} else {
			Status.RunStats[Status.CurrentRun].RejuvPotionsUsed++
		}
	}
}

func ItemStashed(item data.Item, screenshot image.Image) {
	if item.IsPotion() || strings.EqualFold(string(item.Name), "Gold") {
		return
	}

	Status.RunStats[Status.CurrentRun].ItemsFound = append(Status.RunStats[Status.CurrentRun].ItemsFound, item)
	event.Events <- event.Message{
		Message: fmt.Sprintf("Item stashed! %s", item.Name),
		Image:   screenshot,
	}
}

type GameStatus struct {
	ApplicationStartedAt time.Time
	TotalGames           int
	RunStats             map[string]*RunStats

	CurrentRun      string
	CurrentRunStart time.Time
}

type RunStats struct {
	TotalRunsTime          time.Duration
	Kills                  int
	Deaths                 int
	Chickens               int
	MerChicken             int
	Errors                 int
	ItemsFound             []data.Item
	HealingPotionsUsed     int
	ManaPotionsUsed        int
	RejuvPotionsUsed       int
	MercHealingPotionsUsed int
	MercRejuvPotionsUsed   int
}
