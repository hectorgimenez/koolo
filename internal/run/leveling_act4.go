package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
)

func (a Leveling) act4() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		if d.PlayerUnit.Area != area.ThePandemoniumFortress {
			return nil
		}

		return Diablo{baseRun: a.baseRun}.BuildActions()
	})
}
