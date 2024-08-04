package run

import (
	"log/slog"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/container"

	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/health"
)

type Run interface {
	Name() string
	BuildActions() []action.Action
}

type baseRun struct {
	builder *action.Builder
	char    action.Character
	logger  *slog.Logger
	container.Container
}

type Factory struct {
	logger    *slog.Logger
	builder   *action.Builder
	char      action.Character
	container container.Container
	bm        health.BeltManager
}

func NewFactory(logger *slog.Logger, builder *action.Builder, char action.Character, bm health.BeltManager, container container.Container) *Factory {
	return &Factory{
		logger:    logger,
		builder:   builder,
		char:      char,
		bm:        bm,
		container: container,
	}
}

func (f *Factory) BuildRuns() (runs []Run) {
	d := f.container.Reader.GetData(true)

	baseRun := baseRun{
		builder:   f.builder,
		char:      f.char,
		logger:    f.logger,
		Container: f.container,
	}

	if f.container.CharacterCfg.Companion.Enabled && !f.container.CharacterCfg.Companion.Leader {
		return []Run{Companion{baseRun: baseRun}}
	}

	for _, run := range f.container.CharacterCfg.Game.Runs {
		// Prepend terror zone runs, we want to run it always first
		if run == config.TerrorZoneRun {
			tz := TerrorZone{baseRun: baseRun}

			if len(tz.AvailableTZs(d)) > 0 {
				runs = append(runs, tz)
				// If we are skipping other runs, we can return here
				if f.container.CharacterCfg.Game.TerrorZone.SkipOtherRuns {
					return runs
				}
			}
		}
	}

	for _, run := range f.container.CharacterCfg.Game.Runs {
		switch run {
		case config.CountessRun:
			runs = append(runs, Countess{baseRun})
		case config.AndarielRun:
			runs = append(runs, Andariel{baseRun})
		case config.SummonerRun:
			runs = append(runs, Summoner{baseRun})
		case config.DurielRun:
			runs = append(runs, Duriel{baseRun})
		case config.MephistoRun:
			runs = append(runs, Mephisto{baseRun})
		case config.CouncilRun:
			runs = append(runs, Council{baseRun})
		case config.DiabloRun:
			runs = append(runs, Diablo{
				baseRun: baseRun,
				bm:      f.bm,
			})
		case config.EldritchRun:
			runs = append(runs, Eldritch{
				baseRun: baseRun,
			})
		case config.PindleskinRun:
			runs = append(runs, Pindleskin{
				SkipOnImmunities: f.container.CharacterCfg.Game.Pindleskin.SkipOnImmunities,
				baseRun:          baseRun,
			})
		case config.NihlathakRun:
			runs = append(runs, Nihlathak{baseRun})
		case config.AncientTunnelsRun:
			runs = append(runs, AncientTunnels{baseRun})
		case config.MausoleumRun:
			runs = append(runs, Mausoleum{baseRun})
		case config.PitRun:
			runs = append(runs, Pit{baseRun})
		case config.StonyTombRun:
			runs = append(runs, StonyTomb{baseRun})
		case config.ArachnidLairRun:
			runs = append(runs, ArachnidLair{baseRun})
		case config.TristramRun:
			runs = append(runs, Tristram{baseRun})
		case config.LowerKurastRun:
			runs = append(runs, LowerKurast{baseRun})
		case config.LowerKurastChestRun:
			runs = append(runs, LowerKurastChest{baseRun})
		case config.BaalRun:
			runs = append(runs, Baal{baseRun})
		case config.TalRashaTombsRun:
			runs = append(runs, TalRashaTombs{baseRun})
		case config.LevelingRun:
			runs = append(runs, Leveling{baseRun: baseRun, bm: f.bm})
		case config.QuestsRun:
			runs = append(runs, Quests{baseRun})
		case config.CowsRun:
			runs = append(runs, Cows{baseRun})
		case config.ThreshsocketRun:
			runs = append(runs, Threshsocket{baseRun})
		case config.DrifterCavernRun:
			runs = append(runs, DrifterCavern{baseRun})
		case config.EnduguRun:
			runs = append(runs, Endugu{baseRun})
		}
	}

	return
}
