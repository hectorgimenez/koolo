package run

import (
	"log/slog"
	"strings"
	"time"

	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/reader"
)

type Run interface {
	Name() string
	BuildActions() []action.Action
}

type baseRun struct {
	builder *action.Builder
	char    action.Character
	logger  *slog.Logger
}

type Factory struct {
	logger  *slog.Logger
	builder *action.Builder
	char    action.Character
	gr      *reader.GameReader
	bm      health.BeltManager
}

func NewFactory(logger *slog.Logger, builder *action.Builder, char action.Character, gr *reader.GameReader, bm health.BeltManager) *Factory {
	return &Factory{
		logger:  logger,
		builder: builder,
		char:    char,
		gr:      gr,
		bm:      bm,
	}
}

func (f *Factory) BuildRuns() (runs []Run) {
	t := time.Now()
	f.logger.Debug("Fetching map data...")
	d := f.gr.GetData(true)
	f.logger.Debug("Fetch completed", slog.Int64("ms", time.Since(t).Milliseconds()))

	baseRun := baseRun{
		builder: f.builder,
		char:    f.char,
		logger:  f.logger,
	}

	if config.Config.Companion.Enabled && !config.Config.Companion.Leader {
		return []Run{Companion{baseRun: baseRun}}
	}

	for _, run := range config.Config.Game.Runs {
		// Prepend terror zone runs, we want to run it always first
		if run == "terror_zone" {
			tz := TerrorZone{baseRun: baseRun}

			if len(tz.AvailableTZs(d)) > 0 {
				runs = append(runs, tz)
				// If we are skipping other runs, we can return here
				if config.Config.Game.TerrorZone.SkipOtherRuns {
					return runs
				}
			}
		}
	}

	for _, run := range config.Config.Game.Runs {
		run = strings.ToLower(run)
		switch run {
		case "countess":
			runs = append(runs, Countess{baseRun})
		case "andariel":
			runs = append(runs, Andariel{baseRun})
		case "summoner":
			runs = append(runs, Summoner{baseRun})
		case "mephisto":
			runs = append(runs, Mephisto{baseRun})
		case "council":
			runs = append(runs, Council{baseRun})
		case "diablo":
			runs = append(runs, Diablo{
				baseRun: baseRun,
				bm:      f.bm,
			})
		case "eldritch":
			runs = append(runs, Eldritch{
				baseRun: baseRun,
			})
		case "pindleskin":
			runs = append(runs, Pindleskin{
				SkipOnImmunities: config.Config.Game.Pindleskin.SkipOnImmunities,
				baseRun:          baseRun,
			})
		case "nihlathak":
			runs = append(runs, Nihlathak{baseRun})
		case "ancient_tunnels":
			runs = append(runs, AncientTunnels{baseRun})
		case "pit":
			runs = append(runs, Pit{baseRun})
		case "stony_tomb":
			runs = append(runs, StonyTomb{baseRun})
		case "arachnid_lair":
			runs = append(runs, ArachnidLair{baseRun})
		case "tristram":
			runs = append(runs, Tristram{baseRun})
		case "lower_kurast":
			runs = append(runs, LowerKurast{baseRun})
		case "baal":
			runs = append(runs, Baal{baseRun})
		case "tal_rasha_tombs":
			runs = append(runs, TalRashaTombs{baseRun})
		case "leveling":
			runs = append(runs, Leveling{baseRun, f.gr, f.bm})
		case "cows":
			runs = append(runs, Cows{baseRun})
		}
	}

	return
}
