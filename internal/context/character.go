package context

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Character interface {
	CheckKeyBindings() []skill.ID
	BuffSkills() []skill.ID
	PreCTABuffSkills() []skill.ID
	KillCountess() error
	KillAndariel() error
	KillSummoner() error
	KillDuriel() error
	KillMephisto() error
	KillPindle() error
	KillNihlathak() error
	KillCouncil() error
	KillDiablo() error
	KillIzual() error
	KillBaal() error
	KillMonsterSequence(
		monsterSelector func(d game.Data) (data.UnitID, bool),
		skipOnImmunities []stat.Resist,
	) error
}
type StatAllocation struct {
	Stat   stat.ID
	Points int
}

type LevelingCharacter interface {
	Character
	// StatPoints Stats will be assigned in the order they are returned by this function.
	StatPoints() []StatAllocation
	SkillPoints() []skill.ID
	SkillsToBind() (skill.ID, []skill.ID)
	ShouldResetSkills() bool
	KillAncients() error
}
