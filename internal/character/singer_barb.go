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
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	SingerMaxAttackLoops = 15
	singerMaxHorkRange   = 40
)

type SingerBarb struct {
	BaseCharacter
}

func (s SingerBarb) CheckKeyBindings() []skill.ID {
	requiredKeybindings := []skill.ID{skill.WarCry, skill.BattleCommand, skill.TomeOfTownPortal, skill.BattleOrders, skill.Shout, skill.Leap}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requiredKeybindings {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	// Check for one of the armor skills
	if len(missingKeybindings) > 0 {
		s.Logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s SingerBarb) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	previousUnitID := 0
	attackSequenceLoop := 0

	for {
		id, found := monsterSelector(*s.Data)
		if !found {
			s.FindItemOnNearbyCorpses(singerMaxHorkRange)
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
								6,
								false,
								step.Distance(0, 15),
							)
							step.SecondaryAttack(
								skill.WarCry,
								otherMonster.UnitID,
								3,
								step.Distance(0, 3),
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
			} else {
				s.FindItemOnNearbyCorpses(singerMaxHorkRange)
			}
		}

		step.PrimaryAttack(
			monster.UnitID,
			6,
			false,
			step.Distance(0, 15),
		)
		step.SecondaryAttack(
			skill.WarCry,
			monster.UnitID,
			3,
			step.Distance(0, 3),
		)
		action.Buff()

		if attackSequenceLoop >= SingerMaxAttackLoops {
			return nil
		}
		attackSequenceLoop++
		previousUnitID = int(id)
	}
}

func (s SingerBarb) KillBossSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {

	for {
		id, found := monsterSelector(*s.Data)
		if !found {
			s.FindItemOnNearbyCorpses(singerMaxHorkRange)
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

		if monster.Stats[stat.Life] <= 0 {
			s.FindItemOnNearbyCorpses(singerMaxHorkRange)
			return nil
		}

		step.PrimaryAttack(
			monster.UnitID,
			5,
			false,
			step.Distance(0, 15),
		)
		step.SecondaryAttack(
			skill.BattleCry,
			monster.UnitID,
			3,
			step.Distance(0, 3),
		)
		step.SecondaryAttack(
			skill.WarCry,
			monster.UnitID,
			3,
			step.Distance(0, 3),
		)
		action.Buff()

	}
}

func (s SingerBarb) killMonsterByName(id npc.ID, monsterType data.MonsterType, skipOnImmunities []stat.Resist) error {
	return s.KillBossSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}

func (s SingerBarb) BuffSkills() []skill.ID {
	return []skill.ID{skill.BattleCommand, skill.BattleOrders, skill.Shout}
}

func (s SingerBarb) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s SingerBarb) KillAndariel() error {
	return s.killMonsterByName(npc.Andariel, data.MonsterTypeUnique, nil)
}

func (s SingerBarb) KillDuriel() error {
	return s.killMonsterByName(npc.Duriel, data.MonsterTypeUnique, nil)
}

func (s SingerBarb) KillMephisto() error {
	return s.killMonsterByName(npc.Mephisto, data.MonsterTypeUnique, nil)
}

func (s SingerBarb) KillDiablo() error {
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
			if diabloFound {
				return nil
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		diabloFound = true
		s.Logger.Info("Diablo detected, attacking")

		return s.killMonsterByName(npc.Diablo, data.MonsterTypeUnique, nil)
	}
}

func (s SingerBarb) KillBaal() error {
	return s.killMonsterByName(npc.BaalCrab, data.MonsterTypeUnique, nil)
}

func (s SingerBarb) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, nil)
}

func (s SingerBarb) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (s SingerBarb) KillIzual() error {
	return s.killMonsterByName(npc.Izual, data.MonsterTypeUnique, nil)
}

func (s SingerBarb) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		for _, m := range d.Monsters.Enemies() {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				return m.UnitID, true
			}
		}
		return 0, false
	}, nil)
}

func (s SingerBarb) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, s.CharacterCfg.Game.Pindleskin.SkipOnImmunities)
}

func (s SingerBarb) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, nil)
}

func (s *SingerBarb) FindItemOnNearbyCorpses(maxRange int) {
	ctx := context.Get()
	ctx.PauseIfNotPriority()

	findItemKey, found := s.Data.KeyBindings.KeyBindingForSkill(skill.FindItem)
	if !found {
		s.Logger.Debug("Find Item skill not found in key bindings")
		return
	}

	corpses := s.getSortedHorkableCorpses(s.Data.Corpses, maxRange)
	s.Logger.Debug("Horkable corpses found", slog.Int("count", len(corpses)))

	for _, corpse := range corpses {
		err := step.MoveTo(corpse.Position)
		if err != nil {
			s.Logger.Warn("Failed to move to corpse", slog.String("error", err.Error()))
			continue
		}

		if s.Data.PlayerUnit.RightSkill != skill.FindItem {
			ctx.HID.PressKeyBinding(findItemKey)
			time.Sleep(time.Millisecond * 50)
		}

		clickPos := s.getOptimalClickPosition(corpse)
		screenX, screenY := ctx.PathFinder.GameCoordsToScreenCords(clickPos.X, clickPos.Y)
		ctx.HID.Click(game.RightButton, screenX, screenY)
		s.Logger.Debug("Find Item used on corpse", slog.Any("corpse_id", corpse.UnitID))

		time.Sleep(time.Millisecond * 300)
	}

}

func (s *SingerBarb) getOptimalClickPosition(corpse data.Monster) data.Position {
	return data.Position{X: corpse.Position.X, Y: corpse.Position.Y + 1}
}

func (s *SingerBarb) getSortedHorkableCorpses(corpses data.Monsters, maxRange int) []data.Monster {
	var horkableCorpses []data.Monster
	for _, corpse := range corpses {
		if s.isCorpseHorkable(corpse) && s.PathFinder.DistanceFromMe(corpse.Position) <= maxRange {
			horkableCorpses = append(horkableCorpses, corpse)
		}
	}

	sort.Slice(horkableCorpses, func(i, j int) bool {
		distI := s.PathFinder.DistanceFromMe(horkableCorpses[i].Position)
		distJ := s.PathFinder.DistanceFromMe(horkableCorpses[j].Position)
		return distI < distJ
	})

	return horkableCorpses
}

func (s *SingerBarb) isCorpseHorkable(corpse data.Monster) bool {
	unhorkableStates := []state.State{
		state.CorpseNoselect,
		state.CorpseNodraw,
		state.Revive,
		state.Redeemed,
		state.Shatter,
		state.Freeze,
		state.Restinpeace,
	}

	for _, st := range unhorkableStates {
		if corpse.States.HasState(st) {
			return false
		}
	}

	return corpse.Type == data.MonsterTypeChampion ||
		corpse.Type == data.MonsterTypeMinion ||
		corpse.Type == data.MonsterTypeUnique ||
		corpse.Type == data.MonsterTypeSuperUnique
}
