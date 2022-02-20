package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/config"
)

type Run interface {
	Name() string
	BuildActions() []action.Action
}

type baseRun struct {
	builder action.Builder
	char    character.Character
}

func BuildRuns(builder action.Builder, char character.Character) (runs []Run) {
	baseRun := baseRun{
		builder: builder,
		char:    char,
	}

	if config.Config.Runs.Countess {
		runs = append(runs, Countess{baseRun})
	}
	if config.Config.Runs.Andariel {
		runs = append(runs, Andariel{baseRun})
	}
	if config.Config.Runs.Summoner {
		runs = append(runs, Summoner{baseRun})
	}
	if config.Config.Runs.Mephisto {
		runs = append(runs, Mephisto{baseRun})
	}
	if config.Config.Runs.Pindleskin {
		runs = append(runs, Pindleskin{baseRun})
	}

	return
}
