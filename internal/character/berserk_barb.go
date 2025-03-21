package character

import (
	"log/slog"
	"sort"
	"sync/atomic"
	"time"

	"github.com/hectorgimenez/koolo/internal/action"

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

const (
	zerkerMaxHorkRange      = 40
	zerkerMeleeRange        = 5
	zerkerMaxAttackAttempts = 20
)

func (s *Berserker) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.BattleCommand, skill.BattleOrders, skill.Shout, skill.FindItem, skill.Berserk}
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

func (s *Berserker) IsKillingCouncil() bool {
	return s.isKillingCouncil.Load()
}

func (s *Berserker) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {

	for attackAttempts := 0; attackAttempts < zerkerMaxAttackAttempts; attackAttempts++ {
		id, found := monsterSelector(*s.Data)
		if !found {
			if !s.isKillingCouncil.Load() {
				s.FindItemOnNearbyCorpses(zerkerMaxHorkRange)
			}
			return nil
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		monster, monsterFound := s.Data.Monsters.FindByID(id)
		if !monsterFound || monster.Stats[stat.Life] <= 0 {
			continue
		}

		distance := s.PathFinder.DistanceFromMe(monster.Position)
		if distance > zerkerMeleeRange {
			err := step.MoveTo(monster.Position)
			if err != nil {
				s.Logger.Warn("Failed to move to monster", slog.String("error", err.Error()))
				continue
			}
		}

		s.PerformBerserkAttack(monster.UnitID)
		time.Sleep(50 * time.Millisecond)
	}

	return nil
}

func (s *Berserker) PerformBerserkAttack(monsterID data.UnitID) {
	ctx := context.Get()
	ctx.PauseIfNotPriority()
	monster, found := s.Data.Monsters.FindByID(monsterID)
	if !found {
		return
	}

	// Ensure Berserk skill is active
	berserkKey, found := s.Data.KeyBindings.KeyBindingForSkill(skill.Berserk)
	if found && s.Data.PlayerUnit.RightSkill != skill.Berserk {
		ctx.HID.PressKeyBinding(berserkKey)
		time.Sleep(50 * time.Millisecond)
	}

	screenX, screenY := ctx.PathFinder.GameCoordsToScreenCords(monster.Position.X, monster.Position.Y)
	ctx.HID.Click(game.LeftButton, screenX, screenY)
}

func (s *Berserker) FindItemOnNearbyCorpses(maxRange int) {
	ctx := context.Get()
	ctx.PauseIfNotPriority()
	s.SwapToSlot(1)

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

func (s *Berserker) getSortedHorkableCorpses(corpses data.Monsters, maxRange int) []data.Monster {
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

// slot 0 means lowest Gold Find, slot 1 means highest Gold Find
// Presuming attack items will be on slot 0 and Goldfind items on slot 1
// TODO find a way to get active inventory slot from memory.
func (s *Berserker) SwapToSlot(slot int) {
	ctx := context.Get()
	if !ctx.CharacterCfg.Character.BerserkerBarb.FindItemSwitch {
		return // Do nothing if FindItemSwitch is disabled
	}

	initialGF, _ := s.Data.PlayerUnit.FindStat(stat.GoldFind, 0)
	ctx.HID.PressKey('W')
	time.Sleep(100 * time.Millisecond)
	ctx.RefreshGameData()
	swappedGF, _ := s.Data.PlayerUnit.FindStat(stat.GoldFind, 0)

	if (slot == 0 && swappedGF.Value > initialGF.Value) ||
		(slot == 1 && swappedGF.Value < initialGF.Value) {
		ctx.HID.PressKey('W') // Swap back if not in desired slot
	}
}

func (s *Berserker) BuffSkills() []skill.ID {

	skillsList := make([]skill.ID, 0)
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.BattleCommand); found {
		skillsList = append(skillsList, skill.BattleCommand)
	}
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.Shout); found {
		skillsList = append(skillsList, skill.Shout)
	}
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.BattleOrders); found {
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
	return s.killMonster(npc.Andariel, data.MonsterTypeUnique)
}

func (s *Berserker) KillSummoner() error {
	return s.killMonster(npc.Summoner, data.MonsterTypeUnique)
}

func (s *Berserker) KillDuriel() error {
	return s.killMonster(npc.Duriel, data.MonsterTypeUnique)
}

func (s *Berserker) KillMephisto() error {
	return s.killMonster(npc.Mephisto, data.MonsterTypeUnique)
}
func (s *Berserker) KillDiablo() error {
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

		return s.killMonster(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (s *Berserker) KillCouncil() error {
	s.isKillingCouncil.Store(true)
	defer s.isKillingCouncil.Store(false)

	err := s.killAllCouncilMembers()
	if err != nil {
		return err
	}

	context.Get().EnableItemPickup()

	// Wait for corpses to settle
	time.Sleep(500 * time.Millisecond)

	// Perform horking in two passes
	for i := 0; i < 2; i++ {
		s.FindItemOnNearbyCorpses(zerkerMaxHorkRange)

		// Wait between passes
		time.Sleep(300 * time.Millisecond)

		// Refresh game data to catch any new corpses
		context.Get().RefreshGameData()
	}

	// Final wait for items to drop
	time.Sleep(500 * time.Millisecond)

	// Final item pickup
	err = action.ItemPickup(zerkerMaxHorkRange)
	if err != nil {
		s.Logger.Warn("Error during final item pickup after horking", "error", err)
		return err
	}

	// Wait a moment to ensure all items are picked up
	time.Sleep(300 * time.Millisecond)

	return nil
}

func (s *Berserker) killAllCouncilMembers() error {
	context.Get().DisableItemPickup()
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
	for _, m := range s.Data.Monsters.Enemies() {
		if (m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3) && m.Stats[stat.Life] > 0 {
			return true
		}

	}
	return false
}

func (s *Berserker) KillIzual() error {
	return s.killMonster(npc.Izual, data.MonsterTypeUnique)
}

func (s *Berserker) KillPindle() error {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s *Berserker) KillNihlathak() error {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s *Berserker) KillBaal() error {
	return s.killMonster(npc.BaalCrab, data.MonsterTypeUnique)
}
