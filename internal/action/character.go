package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
)

type Character interface {
	Buff() Action
	KillCountess() Action
	KillAndariel() Action
	KillSummoner() Action
	KillMephisto() Action
	KillPindle(skipOnImmunities []stat.Resist) Action
	KillNihlathak() Action
	KillCouncil() Action
	KillMonsterSequence(
		monsterSelector func(d data.Data) (data.UnitID, bool),
		skipOnImmunities []stat.Resist,
		opts ...step.AttackOption,
	) *DynamicAction
}
