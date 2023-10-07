package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
)

type Character interface {
	BuffSkills() map[skill.Skill]string
	KillCountess() Action
	KillAndariel() Action
	KillSummoner() Action
	KillDuriel() Action
	KillMephisto() Action
	KillPindle(skipOnImmunities []stat.Resist) Action
	KillNihlathak() Action
	KillCouncil() Action
	KillDiablo() Action
	KillIzual() Action
	KillBaal() Action
	KillMonsterSequence(
		monsterSelector func(d data.Data) (data.UnitID, bool),
		skipOnImmunities []stat.Resist,
		opts ...step.AttackOption,
	) Action
}

type LevelingCharacter interface {
	Character
	// StatPoints Stats will be assigned in the order they are returned by this function.
	StatPoints(data.Data) map[stat.ID]int
	SkillPoints(data.Data) []skill.Skill
	GetKeyBindings(data.Data) map[skill.Skill]string
	ShouldResetSkills(data.Data) bool
	KillAncients() Action
	GetSkillTree() skill.Tree
}
