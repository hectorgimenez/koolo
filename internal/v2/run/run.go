package run

import "github.com/hectorgimenez/koolo/internal/config"

type Run interface {
	Name() string
	Run() error
}

func BuildRuns(cfg *config.CharacterCfg) (runs []Run) {
	//if cfg.Companion.Enabled && !cfg.Companion.Leader {
	//	return []Run{Companion{baseRun: baseRun}}
	//}
	//
	//for _, run := range f.container.CharacterCfg.Game.Runs {
	//	// Prepend terror zone runs, we want to run it always first
	//	if run == config.TerrorZoneRun {
	//		tz := TerrorZone{baseRun: baseRun}
	//
	//		if len(tz.AvailableTZs(d)) > 0 {
	//			runs = append(runs, tz)
	//			// If we are skipping other runs, we can return here
	//			if f.container.CharacterCfg.Game.TerrorZone.SkipOtherRuns {
	//				return runs
	//			}
	//		}
	//	}
	//}

	for _, run := range cfg.Game.Runs {
		switch run {
		case config.CountessRun:
			runs = append(runs, NewCountess())
		case config.AndarielRun:
			runs = append(runs, NewAndariel())
		//case config.SummonerRun:
		//runs = append(runs, Summoner{baseRun})
		//case config.DurielRun:
		//runs = append(runs, Duriel{baseRun})
		//case config.MephistoRun:
		//runs = append(runs, Mephisto{baseRun})
		case config.TravincalRun:
			runs = append(runs, NewTravincal())
		//case config.DiabloRun:
		//runs = append(runs, Diablo{
		//	baseRun: baseRun,
		//	bm:      f.bm,
		//})
		//case config.EldritchRun:
		//runs = append(runs, Eldritch{
		//	baseRun: baseRun,
		//})
		case config.PindleskinRun:
			runs = append(runs, NewPindleskin())
		//case config.NihlathakRun:
		//runs = append(runs, Nihlathak{baseRun})
		case config.AncientTunnelsRun:
			runs = append(runs, NewAncientTunnels())
		//case config.MausoleumRun:
		//	runs = append(runs, Mausoleum{baseRun})
		//case config.PitRun:
		//	runs = append(runs, Pit{baseRun})
		//case config.StonyTombRun:
		//	runs = append(runs, StonyTomb{baseRun})
		case config.ArachnidLairRun:
			runs = append(runs, NewArachnidLair())
		//case config.TristramRun:
		//	runs = append(runs, Tristram{baseRun})
		//case config.LowerKurastRun:
		//	runs = append(runs, LowerKurast{baseRun})
		//case config.LowerKurastChestRun:
		//	runs = append(runs, LowerKurastChest{baseRun})
		case config.BaalRun:
			runs = append(runs, NewBaal())
			//case config.TalRashaTombsRun:
			//	runs = append(runs, TalRashaTombs{baseRun})
			//case config.LevelingRun:
			//	runs = append(runs, Leveling{baseRun: baseRun, bm: f.bm})
			//case config.QuestsRun:
			//	runs = append(runs, Quests{baseRun})
			//case config.CowsRun:
			//	runs = append(runs, Cows{baseRun})
			//case config.ThreshsocketRun:
			//	runs = append(runs, Threshsocket{baseRun})
			//case config.DrifterCavernRun:
			//	runs = append(runs, DrifterCavern{baseRun})
			//case config.EnduguRun:
			//	runs = append(runs, Endugu{baseRun})
		}
	}

	return runs
}
