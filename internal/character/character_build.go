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
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

const ()

type CharacterBuild struct {
	BaseCharacter
}

func (s CharacterBuild) CheckKeyBindings() []skill.ID {
	missingKeybindings := []skill.ID{}
	return missingKeybindings
}

func (s CharacterBuild) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	completedAttackLoops := 0
	previousUnitID := 0
	previousSelfBlizzard := time.Time{}

	blizzOpts := step.StationaryDistance(blizzMinDistance, blizzMaxDistance)
	lsOpts := step.Distance(LSMinDistance, LSMaxDistance)

	for {
		id, found := monsterSelector(*s.Data)
		if !found {
			return nil
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		if completedAttackLoops >= sorceressMaxAttacksLoop {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		// Cast a Blizzard on very close mobs, in order to clear possible trash close the player, every two attack rotations
		if time.Since(previousSelfBlizzard) > time.Second*4 && !s.Data.PlayerUnit.States.HasState(state.Cooldown) {
			for _, m := range s.Data.Monsters.Enemies() {
				if dist := s.PathFinder.DistanceFromMe(m.Position); dist < 4 {
					previousSelfBlizzard = time.Now()
					step.SecondaryAttack(skill.Blizzard, m.UnitID, 1, blizzOpts)
				}
			}
		}

		if s.Data.PlayerUnit.States.HasState(state.Cooldown) {
			step.PrimaryAttack(id, 2, true, lsOpts)
		}

		step.SecondaryAttack(skill.Blizzard, id, 1, blizzOpts)

		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func (s CharacterBuild) killMonster(npc npc.ID, t data.MonsterType) error {
	// Check to see if static field has a hotkey
	staticField := []skill.ID{skill.StaticField}[0]
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(staticField); found {
		m, _ := s.Data.Monsters.FindOne(npc, t)
		step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(5, 8))
	}

	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s CharacterBuild) killMonsterByName(id npc.ID, monsterType data.MonsterType, skipOnImmunities []stat.Resist) error {
	// Check to see if static field has a hotkey
	staticField := []skill.ID{skill.StaticField}[0]
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(staticField); found {
		m, _ := s.Data.Monsters.FindOne(id, monsterType)
		step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(5, 8))
	}

	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}

func (s CharacterBuild) BuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s CharacterBuild) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s CharacterBuild) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, nil)
}

func (s CharacterBuild) KillAndariel() error {
	return s.killMonster(npc.Andariel, data.MonsterTypeUnique)
}

func (s CharacterBuild) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (s CharacterBuild) KillDuriel() error {
	return s.killMonsterByName(npc.Duriel, data.MonsterTypeUnique, nil)
}

func (s CharacterBuild) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		var councilMembers []data.Monster
		for _, m := range d.Monsters {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				councilMembers = append(councilMembers, m)
			}
		}

		// Order council members by distance
		sort.Slice(councilMembers, func(i, j int) bool {
			distanceI := s.PathFinder.DistanceFromMe(councilMembers[i].Position)
			distanceJ := s.PathFinder.DistanceFromMe(councilMembers[j].Position)

			return distanceI < distanceJ
		})

		if len(councilMembers) > 0 {
			s.Logger.Debug("Targeting Council member", "id", councilMembers[0].UnitID)
			return councilMembers[0].UnitID, true
		}

		s.Logger.Debug("No Council members found")
		return 0, false
	}, nil)
}

func (s CharacterBuild) KillMephisto() error {
	return s.killMonsterByName(npc.Mephisto, data.MonsterTypeUnique, nil)
}

func (s CharacterBuild) KillIzual() error {
	return s.killMonster(npc.Izual, data.MonsterTypeUnique)
}

func (s CharacterBuild) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			s.Logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := s.Data.Monsters.FindOne(npc.Diablo, data.MonsterTypeUnique)
		if !found || diablo.Stats[stat.Life] <= 0 {
			// Already dead
			if diabloFound {
				return nil
			}

			// Keep waiting...
			time.Sleep(200 * time.Millisecond)
			continue
		}

		diabloFound = true
		s.Logger.Info("Diablo detected, attacking")

		return s.killMonster(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (s CharacterBuild) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, s.CharacterCfg.Game.Pindleskin.SkipOnImmunities)
}

func (s CharacterBuild) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, nil)
}

func (s CharacterBuild) KillBaal() error {
	return s.killMonster(npc.BaalCrab, data.MonsterTypeUnique)
}

func (s CharacterBuild) KillAncients() error {
	for _, m := range s.Data.Monsters.Enemies(data.MonsterEliteFilter()) {
		m, _ := s.Data.Monsters.FindOne(m.Name, data.MonsterTypeSuperUnique)

		s.killMonster(m.Name, data.MonsterTypeSuperUnique)
	}
	return nil
}
