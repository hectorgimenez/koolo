package character

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	PsnNovaMinDistance    = 6
	PsnNovaMaxDistance    = 12
	CurseMinDistance      = 6
	CurseMaxDistance      = 25
	PsnNovaMaxAttacksLoop = 30
)

type PoisonNecro struct {
	BaseCharacter
}

func (p PoisonNecro) CheckKeyBindings() []skill.ID {
	requiredKeybindings := []skill.ID{skill.PoisonNova, skill.Teleport, skill.TomeOfTownPortal, skill.LowerResist}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requiredKeybindings {
		if _, found := p.Data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		p.Logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (p PoisonNecro) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	ctx := context.Get()
	completedAttackLoops := 0
	previousUnitID := 0

	curseOpts := []step.AttackOption{
		step.RangedDistance(CurseMinDistance, CurseMaxDistance),
	}
	psnNovaOpts := []step.AttackOption{
		step.RangedDistance(PsnNovaMinDistance, PsnNovaMaxDistance),
	}

	for {
		ctx.PauseIfNotPriority()

		id, found := monsterSelector(*p.Data)
		if !found {
			return nil
		}

		if !p.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		monster, found := p.Data.Monsters.FindByID(id)
		if !found || monster.Stats[stat.Life] <= 0 {
			return nil
		}

		// Cast Lower Resist first
		_ = step.SecondaryAttack(skill.LowerResist, monster.UnitID, 1, curseOpts...)

		// If area is unreachable, or monster is dead, skip.
		if previousUnitID == int(id) {
			if monster.Stats[stat.Life] > 0 {
				if p.Data.AreaData.IsWalkable(monster.Position) {
					ctx := context.Get()
					otherMonsterLoopCounter := 0
					for _, otherMonster := range ctx.Data.Monsters.Enemies() {
						if otherMonster.Stats[stat.Life] > 0 && pather.DistanceFromPoint(p.Data.PlayerUnit.Position, otherMonster.Position) <= 30 && ctx.Data.AreaData.IsWalkable(otherMonster.Position) {
							otherMonsterLoopCounter++
							_ = step.SecondaryAttack(skill.LowerResist, otherMonster.UnitID, 1, curseOpts...)
							_ = step.SecondaryAttack(skill.PoisonNova, otherMonster.UnitID, 6, psnNovaOpts...)
						}
					}
					if otherMonsterLoopCounter == 0 {
						p.PathFinder.RandomTeleport() // will walk if can't teleport
						utils.Sleep(400)
					}
				} else {
					continue
				}
			}
		}

		_ = step.SecondaryAttack(skill.PoisonNova, monster.UnitID, 6, psnNovaOpts...)
		completedAttackLoops++
	}
}

func (p PoisonNecro) KillBossSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	previousUnitID := 0
	curseOpts := []step.AttackOption{
		step.RangedDistance(CurseMinDistance, CurseMaxDistance),
	}
	psnNovaOpts := []step.AttackOption{
		step.RangedDistance(PsnNovaMinDistance, PsnNovaMaxDistance),
	}

	for {
		id, found := monsterSelector(*p.Data)
		if !found {
			return nil
		}

		monster, found := p.Data.Monsters.FindByID(id)
		if !found {
			p.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		if previousUnitID == int(id) {
			if monster.Stats[stat.Life] > 0 {
				p.PathFinder.RandomTeleport() // will walk if can't teleport
				utils.Sleep(400)
				action.MoveToCoords(data.Position{X: monster.Position.X - 2, Y: monster.Position.Y - 2})
			} else {
				return nil
			}
		}
		_ = step.SecondaryAttack(skill.LowerResist, monster.UnitID, 1, curseOpts...)
		_ = step.SecondaryAttack(skill.PoisonNova, monster.UnitID, 6, psnNovaOpts...)

		previousUnitID = int(id)
	}
}

func (p PoisonNecro) killMonsterByName(id npc.ID, monsterType data.MonsterType, skipOnImmunities []stat.Resist) error {
	return p.KillBossSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}

func (p PoisonNecro) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := p.Data.KeyBindings.KeyBindingForSkill(skill.BoneArmor); found {
		skillsList = append(skillsList, skill.BoneArmor)
	}
	if _, found := p.Data.KeyBindings.KeyBindingForSkill(skill.ClayGolem); found {
		skillsList = append(skillsList, skill.ClayGolem)
	}

	return skillsList
}

func (p PoisonNecro) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (p PoisonNecro) KillAndariel() error {
	return p.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (p PoisonNecro) KillDuriel() error {
	return p.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (p PoisonNecro) KillMephisto() error {
	return p.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (p PoisonNecro) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			p.Logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := p.Data.Monsters.FindOne(npc.Diablo, data.MonsterTypeUnique)
		if !found || diablo.Stats[stat.Life] <= 0 {
			if diabloFound {
				return nil
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		diabloFound = true
		p.Logger.Info("Diablo detected, attacking")

		return p.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
	}
}

func (p PoisonNecro) KillBaal() error {
	return p.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (p PoisonNecro) KillCountess() error {
	return p.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, nil)
}

func (p PoisonNecro) KillSummoner() error {
	return p.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (p PoisonNecro) KillIzual() error {
	return p.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (p PoisonNecro) KillCouncil() error {
	return p.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		for _, m := range d.Monsters.Enemies() {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				return m.UnitID, true
			}
		}
		return 0, false
	}, nil)
}

func (p PoisonNecro) KillPindle() error {
	return p.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, p.CharacterCfg.Game.Pindleskin.SkipOnImmunities)
}

func (p PoisonNecro) KillNihlathak() error {
	return p.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, nil)
}
