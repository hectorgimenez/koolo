package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Character interface {
	CheckKeyBindings(game.Data) []skill.ID
	BuffSkills(game.Data) []skill.ID
	PreCTABuffSkills(game.Data) []skill.ID
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
		monsterSelector func(d game.Data) (data.UnitID, bool),
		skipOnImmunities []stat.Resist,
		opts ...step.AttackOption,
	) Action
}

type LevelingCharacter interface {
	Character
	// StatPoints Stats will be assigned in the order they are returned by this function.
	StatPoints(game.Data) map[stat.ID]int
	SkillPoints(game.Data) []skill.ID
	SkillsToBind(game.Data) (skill.ID, []skill.ID)
	ShouldResetSkills(game.Data) bool
	KillAncients() Action
}
