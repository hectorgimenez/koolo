package character

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/utils"
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
)

const (
	hammerdinMaxAttacksLoop = 20 // Adjust from 5-20 depending on DMG and rotation, lower attack loops would cause higher attack rotation whereas bigger would perform multiple(longer) attacks on one spot.
)

type Hammerdin struct {
	BaseCharacter
}

func (s Hammerdin) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.Concentration, skill.HolyShield, skill.TomeOfTownPortal}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		s.Logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s Hammerdin) KillBossSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	previousUnitID := 0

	for {
		id, found := monsterSelector(*s.Data)
		if !found {
			return nil
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		// Reposition right under for perfect hammer hit
		if previousUnitID == int(id) {
			if monster.Stats[stat.Life] > 0 {
				s.PathFinder.RandomTeleport() // will walk if can't teleport
				utils.Sleep(400)
				action.MoveToCoords(data.Position{monster.Position.X - 2, monster.Position.Y - 2})
			}
		}

		step.PrimaryAttack(
			id,
			4,
			true,
			step.Distance(2, 2), // X,Y coords of 2,2 is the perfect hammer angle attack for NPC targeting/attacking, you can adjust accordingly anything between 1,1 - 3,3 is acceptable, where the higher the number, the bigger the distance from the player (usually used for De Seis)
			step.EnsureAura(skill.Concentration),
		)

		previousUnitID = int(id)
	}
}

func (s Hammerdin) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	previousUnitID := 0
	attackSequenceLoop := 0

	for {
		id, found := monsterSelector(*s.Data)
		if !found {
			return nil
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		// If area is unreachable, or monster is dead, skip.
		if previousUnitID == int(id) {
			if monster.Stats[stat.Life] > 0 {
				if s.Data.AreaData.IsWalkable(monster.Position) {
					ctx := context.Get()
					otherMonsterLoopCounter := 0
					for _, otherMonster := range ctx.Data.Monsters.Enemies() {
						if otherMonster.Stats[stat.Life] > 0 && pather.DistanceFromPoint(s.Data.PlayerUnit.Position, otherMonster.Position) <= 30 && ctx.Data.AreaData.IsWalkable(otherMonster.Position) {
							otherMonsterLoopCounter++
							step.PrimaryAttack(
								otherMonster.UnitID,
								4,
								true,
								step.Distance(2, 2), // X,Y coords of 2,2 is the perfect hammer angle attack for NPC targeting/attacking, you can adjust accordingly anything between 1,1 - 3,3 is acceptable, where the higher the number, the bigger the distance from the player (usually used for De Seis)
								step.EnsureAura(skill.Concentration),
							)
						}
					}
					if otherMonsterLoopCounter == 0 {
						s.PathFinder.RandomTeleport() // will walk if can't teleport
						utils.Sleep(400)
					}
				} else {
					continue
				}
			}
		}

		step.PrimaryAttack(
			id,
			4,
			true,
			step.Distance(2, 2), // X,Y coords of 2,2 is the perfect hammer angle attack for NPC targeting/attacking, you can adjust accordingly anything between 1,1 - 3,3 is acceptable, where the higher the number, the bigger the distance from the player (usually used for De Seis)
			step.EnsureAura(skill.Concentration),
		)

		if attackSequenceLoop >= hammerdinMaxAttacksLoop {
			return nil
		}
		attackSequenceLoop++
		previousUnitID = int(id)
	}
}

func (s Hammerdin) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s Hammerdin) killMonsterByName(id npc.ID, monsterType data.MonsterType, _ bool) error {
	return s.KillBossSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (s Hammerdin) BuffSkills() []skill.ID {

	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.HolyShield); found {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.OakSage); found {

			return []skill.ID{skill.HolyShield, skill.OakSage}
		}
		return []skill.ID{skill.HolyShield}
	}

	return []skill.ID{}
}

func (s Hammerdin) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s Hammerdin) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, false)
}

func (s Hammerdin) KillAndariel() error {
	return s.killMonsterByName(npc.Andariel, data.MonsterTypeUnique, false)
}
func (s Hammerdin) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, false)
}

func (s Hammerdin) KillDuriel() error {
	return s.killMonsterByName(npc.Duriel, data.MonsterTypeUnique, false)
}

func (s Hammerdin) KillCouncil() error {
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
			distanceI := s.PathFinder.DistanceFromMe(councilMembers[i].Position)
			distanceJ := s.PathFinder.DistanceFromMe(councilMembers[j].Position)

			return distanceI < distanceJ
		})

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (s Hammerdin) KillMephisto() error {
	return s.killMonsterByName(npc.Mephisto, data.MonsterTypeUnique, false)
}
func (s Hammerdin) KillIzual() error {
	return s.killMonsterByName(npc.Izual, data.MonsterTypeUnique, false)
}

func (s Hammerdin) KillDiablo() error {
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
			time.Sleep(200)
			continue
		}

		diabloFound = true
		s.Logger.Info("Diablo detected, attacking")

		return s.killMonsterByName(npc.Diablo, data.MonsterTypeUnique, false)
	}
}

func (s Hammerdin) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, false)
}

func (s Hammerdin) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, false)
}

func (s Hammerdin) KillBaal() error {
	return s.killMonsterByName(npc.BaalCrab, data.MonsterTypeUnique, false)
}
