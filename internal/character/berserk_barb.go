package character

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"log/slog"
	"sort"
	"sync/atomic"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Berserker struct {
	BaseCharacter
	isKillingCouncil atomic.Bool
}

func (s *Berserker) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.BattleCommand, skill.BattleOrders, skill.Shout, skill.FindItem, skill.TomeOfTownPortal}
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

func (s *Berserker) IsKillingCouncil() bool {
	return s.isKillingCouncil.Load()
}

func (s *Berserker) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	attackAttempts := 0
	maxAttackAttempts := 5
	const maxRange = 30

	for {
		id, found := monsterSelector(*s.data)
		if !found {
			// If no monsters are found and we're not killing council, attempt to hork
			if !s.isKillingCouncil.Load() {
				s.FindItemOnNearbyCorpses(maxRange)
			}
			return nil
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		monster, monsterFound := s.data.Monsters.FindByID(id)
		if !monsterFound {
			return nil
		}

		opts := []step.AttackOption{step.Distance(1, maxRange)}

		if attackAttempts >= maxAttackAttempts {
			return nil
		}

		err := step.MoveTo(monster.Position)
		if err != nil {
			s.logger.Warn("Failed to move to monster", slog.String("error", err.Error()))
			continue
		}

		err = step.PrimaryAttack(id, 1, false, opts...)
		if err != nil {
			s.logger.Warn("Failed to attack monster", slog.String("error", err.Error()))
		}

		attackAttempts++
	}
}

/*  killAndHork  for later use  like
func (s *Berserker) KillPindle() error {
    return s.killAndHork(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}*/

func (s *Berserker) killAndHork(npc npc.ID, t data.MonsterType) error {
	err := s.killMonster(npc, t)
	if err != nil {
		return err
	}
	s.FindItemOnNearbyCorpses(30)
	return action.ItemPickup(30)
}

func (s *Berserker) FindItemOnNearbyCorpses(maxRange int) {
	ctx := context.Get()

	findItemKey, found := s.data.KeyBindings.KeyBindingForSkill(skill.FindItem)
	if !found {
		s.logger.Debug("Find Item skill not found in key bindings")
		return
	}

	corpses := s.getSortedHorkableCorpses(s.data.Corpses, s.data.PlayerUnit.Position, maxRange)
	s.logger.Debug("Horkable corpses found", slog.Int("count", len(corpses)))

	for _, corpse := range corpses {
		err := step.MoveTo(corpse.Position)
		if err != nil {
			s.logger.Warn("Failed to move to corpse", slog.String("error", err.Error()))
			continue
		}

		if s.data.PlayerUnit.RightSkill != skill.FindItem {
			ctx.HID.PressKeyBinding(findItemKey)
			time.Sleep(time.Millisecond * 50)
		}

		clickPos := s.getOptimalClickPosition(corpse)
		screenX, screenY := ctx.PathFinder.GameCoordsToScreenCords(clickPos.X, clickPos.Y)
		ctx.HID.Click(game.RightButton, screenX, screenY)
		s.logger.Debug("Find Item used on corpse", slog.Any("corpse_id", corpse.UnitID))

		time.Sleep(time.Millisecond * 500)
	}
}

func (s *Berserker) getSortedHorkableCorpses(corpses data.Monsters, playerPos data.Position, maxRange int) []data.Monster {
	var horkableCorpses []data.Monster
	for _, corpse := range corpses {
		if s.isCorpseHorkable(corpse) && s.pf.DistanceFromMe(corpse.Position) <= maxRange {
			horkableCorpses = append(horkableCorpses, corpse)
		}
	}

	sort.Slice(horkableCorpses, func(i, j int) bool {
		distI := s.pf.DistanceFromMe(horkableCorpses[i].Position)
		distJ := s.pf.DistanceFromMe(horkableCorpses[j].Position)
		return distI < distJ
	})

	return horkableCorpses
}

func (s *Berserker) isCorpseHorkable(corpse data.Monster) bool {
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

func (s *Berserker) getOptimalClickPosition(corpse data.Monster) data.Position {
	return data.Position{X: corpse.Position.X, Y: corpse.Position.Y + 1}
}

func (s *Berserker) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.BattleCommand); found {
		skillsList = append(skillsList, skill.BattleCommand)
	}
	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.Shout); found {
		skillsList = append(skillsList, skill.Shout)
	}
	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.BattleOrders); found {
		skillsList = append(skillsList, skill.BattleOrders)
	}
	return skillsList
}

func (s *Berserker) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s *Berserker) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}
		return m.UnitID, true
	}, nil)
}

func (s *Berserker) KillCountess() error {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s *Berserker) KillAndariel() error {
	return s.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (s *Berserker) KillSummoner() error {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s *Berserker) KillDuriel() error {
	return s.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (s *Berserker) KillMephisto() error {
	return s.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (s *Berserker) KillDiablo() error {
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
			if diabloFound {
				return nil
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		diabloFound = true
		s.logger.Info("Diablo detected, attacking")

		return s.killMonster(npc.Diablo, data.MonsterTypeNone)
	}
}

func (s *Berserker) KillCouncil() error {
	s.isKillingCouncil.Store(true)
	defer s.isKillingCouncil.Store(false)

	// First, kill all council members
	err := s.killAllCouncilMembers()
	if err != nil {
		return err
	}

	// Then, hork corpses and pickup items
	s.FindItemOnNearbyCorpses(30)
	return action.ItemPickup(30)

}

func (s *Berserker) killAllCouncilMembers() error {
	for {
		if !s.anyCouncilMemberAlive() {
			return nil
		}

		err := s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			for _, m := range d.Monsters.Enemies() {
				if (m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3) && m.Stats[stat.Life] > 0 {
					return m.UnitID, true
				}
			}
			return 0, false
		}, nil)

		if err != nil {
			return err
		}
	}
}

func (s *Berserker) anyCouncilMemberAlive() bool {
	for _, m := range s.data.Monsters.Enemies() {
		if (m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3) && m.Stats[stat.Life] > 0 {
			return true
		}
	}
	return false
}

func (s *Berserker) KillIzual() error {
	return s.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (s *Berserker) KillPindle() error {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s *Berserker) KillNihlathak() error {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s *Berserker) KillBaal() error {
	return s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}
