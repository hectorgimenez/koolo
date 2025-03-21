package run

import (
	"github.com/hectorgimenez/koolo/internal/config"
)

type Run interface {
	Name() string
	Run() error
}

func BuildRuns(cfg *config.CharacterCfg) (runs []Run) {
	//if cfg.Companion.Enabled && !cfg.Companion.Leader {
	//	return []Run{Companion{baseRun: baseRun}}
	//}

	for _, run := range cfg.Game.Runs {
		// Prepend terror zone runs, we want to run it always first
		if run == config.TerrorZoneRun {
			tz := NewTerrorZone()

			if len(tz.AvailableTZs()) > 0 {
				runs = append(runs, tz)
				// If we are skipping other runs, we can return here
				if cfg.Game.TerrorZone.SkipOtherRuns {
					return runs
				}
			}
		}
	}

	for _, run := range cfg.Game.Runs {
		switch run {
		case config.CountessRun:
			runs = append(runs, NewCountess())
		case config.AndarielRun:
			runs = append(runs, NewAndariel())
		case config.SummonerRun:
			runs = append(runs, NewSummoner())
		case config.DurielRun:
			runs = append(runs, NewDuriel())
		case config.MephistoRun:
			runs = append(runs, NewMephisto(nil))
		case config.TravincalRun:
			runs = append(runs, NewTravincal())
		case config.DiabloRun:
			runs = append(runs, NewDiablo())
		case config.EldritchRun:
			runs = append(runs, NewEldritch())
		case config.PindleskinRun:
			runs = append(runs, NewPindleskin())
		case config.NihlathakRun:
			runs = append(runs, NewNihlathak())
		case config.AncientTunnelsRun:
			runs = append(runs, NewAncientTunnels())
		case config.MausoleumRun:
			runs = append(runs, NewMausoleum())
		case config.PitRun:
			runs = append(runs, NewPit())
		case config.StonyTombRun:
			runs = append(runs, NewStonyTomb())
		case config.ArachnidLairRun:
			runs = append(runs, NewArachnidLair())
		case config.TristramRun:
			runs = append(runs, NewTristram())
		case config.LowerKurastRun:
			runs = append(runs, NewLowerKurast())
		case config.LowerKurastChestRun:
			runs = append(runs, NewLowerKurastChest())
		case config.BaalRun:
			runs = append(runs, NewBaal(nil))
		case config.TalRashaTombsRun:
			runs = append(runs, NewTalRashaTombs())
		case config.LevelingRun:
			runs = append(runs, NewLeveling())
		case config.QuestsRun:
			runs = append(runs, NewQuests())
		case config.FollowerRun:
			runs = append(runs, NewFollower())
		case config.CowsRun:
			runs = append(runs, NewCows())
		case config.ThreshsocketRun:
			runs = append(runs, NewThreshsocket())
		case config.SpiderCavernRun:
			runs = append(runs, NewSpiderCavern())
		case config.DrifterCavernRun:
			runs = append(runs, NewDriverCavern())
		case config.EnduguRun:
			runs = append(runs, NewEndugu())
		}
	}

	return runs
}
