package run

import (
	"strings"

	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/reader"
	"go.uber.org/zap"
)

type Run interface {
	Name() string
	BuildActions() []action.Action
}

type baseRun struct {
	builder action.Builder
	char    action.Character
	logger  *zap.Logger
}

func BuildRuns(logger *zap.Logger, builder action.Builder, char action.Character, gr *reader.GameReader, bm health.BeltManager) (runs []Run) {
	baseRun := baseRun{
		builder: builder,
		char:    char,
		logger:  logger,
	}

	if config.Config.Companion.Enabled && !config.Config.Companion.Leader {
		return []Run{Companion{baseRun: baseRun}}
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
				bm:      bm,
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
		case "leveling":
			runs = append(runs, Leveling{baseRun, gr, bm})
		}
	}

	return
}
