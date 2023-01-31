package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"go.uber.org/zap"
	"strings"
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

func BuildRuns(logger *zap.Logger, builder action.Builder, char action.Character) (runs []Run) {
	baseRun := baseRun{
		builder: builder,
		char:    char,
		logger:  logger,
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
		case "tristram":
			runs = append(runs, Tristram{baseRun})
		case "lower_kurast":
			runs = append(runs, LowerKurast{baseRun})
		}
	}

	return
}
