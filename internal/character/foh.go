package character

import (
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	fohMaxAttacksLoop = 20
	fohMinDistance    = 5
	fohMaxDistance    = 20
)

type Foh struct {
	BaseCharacter
	*game.HID
}

func (s Foh) CheckKeyBindings() []skill.ID {

	requireKeybindings := []skill.ID{skill.Conviction, skill.HolyShield, skill.TomeOfTownPortal}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := s.data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		s.logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s Foh) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	completedAttackLoops := 0
	previousUnitID := 0

	for {
		id, found := monsterSelector(*s.data)
		if !found {
			return nil
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		if completedAttackLoops >= fohMaxAttacksLoop {
			return nil
		}

		monster, found := s.data.Monsters.FindByID(id)
		if !found {
			s.logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		if s.data.PlayerUnit.LeftSkill != skill.FistOfTheHeavens {
			fohKey, fohFound := s.data.KeyBindings.KeyBindingForSkill(skill.FistOfTheHeavens)
			if fohFound {
				utils.Sleep(40)
				s.HID.PressKeyBinding(fohKey)
			}
		}

		step.PrimaryAttack(
			id,
			3,
			true,
			step.Distance(fohMinDistance, fohMaxDistance),
			step.EnsureAura(skill.Conviction),
		)

		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func (s Foh) killMonsterByName(id npc.ID, monsterType data.MonsterType, _ bool) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (s Foh) killBoss(id npc.ID, monsterType data.MonsterType) error {
	for {
		monster, found := s.data.Monsters.FindOne(id, monsterType)
		if !found || monster.Stats[stat.Life] <= 0 {
			utils.Sleep(100)
			return nil
		}

		hbKey, holyBoltFound := s.data.KeyBindings.KeyBindingForSkill(skill.HolyBolt)
		fohKey, fohFound := s.data.KeyBindings.KeyBindingForSkill(skill.FistOfTheHeavens)

		// Switch between foh and holy bolt while attacking
		if holyBoltFound && fohFound {
			utils.Sleep(50)

			step.PrimaryAttack(
				monster.UnitID,
				1,
				true,
				step.Distance(fohMinDistance, fohMaxDistance),
				step.EnsureAura(skill.Conviction),
			)
			s.HID.PressKeyBinding(hbKey)
			utils.Sleep(40)

			step.PrimaryAttack(
				monster.UnitID,
				3,
				true,
				step.Distance(fohMinDistance, fohMaxDistance),
				step.EnsureAura(skill.Conviction),
			)

			utils.Sleep(40)
			s.HID.PressKeyBinding(fohKey)
		} else {
			utils.Sleep(100)
			// Don't switch because no keybindings found
			step.PrimaryAttack(
				monster.UnitID,
				3,
				true,
				step.Distance(fohMinDistance, fohMaxDistance),
				step.EnsureAura(skill.Conviction),
			)
		}
	}
}

func (s Foh) BuffSkills() []skill.ID {
	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.HolyShield); found {
		return []skill.ID{skill.HolyShield}
	}
	return []skill.ID{}
}

func (s Foh) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s Foh) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, false)
}

func (s Foh) KillAndariel() error {
	return s.killBoss(npc.Andariel, data.MonsterTypeNone)
}

func (s Foh) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeNone, false)
}

func (s Foh) KillDuriel() error {
	return s.killBoss(npc.Duriel, data.MonsterTypeNone)
}

func (s Foh) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		// Exclude monsters that are not council members
		var councilMembers []data.Monster
		for _, m := range d.Monsters {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				councilMembers = append(councilMembers, m)
			}
		}

		// Order council members by distance
		sort.Slice(councilMembers, func(i, j int) bool {
			distanceI := s.pf.DistanceFromMe(councilMembers[i].Position)
			distanceJ := s.pf.DistanceFromMe(councilMembers[j].Position)

			return distanceI < distanceJ
		})

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (s Foh) KillMephisto() error {
	return s.killBoss(npc.Mephisto, data.MonsterTypeNone)
}

func (s Foh) KillIzual() error {
	return s.killMonsterByName(npc.Izual, data.MonsterTypeNone, false)
}

func (s Foh) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			s.logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := s.data.Monsters.FindOne(npc.Diablo, data.MonsterTypeNone)
		if !found || diablo.Stats[stat.Life] <= 0 {
			// Already dead
			if diabloFound {
				return nil
			}

			// Keep waiting...
			time.Sleep(200)
			continue
		}

		diabloFound = true
		s.logger.Info("Diablo detected, attacking")

		return s.killBoss(npc.Diablo, data.MonsterTypeNone)
	}
}

func (s Foh) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, false)
}

func (s Foh) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, false)
}

func (s Foh) KillBaal() error {
	return s.killBoss(npc.BaalCrab, data.MonsterTypeNone)
}
