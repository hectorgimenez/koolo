package character

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data/mode"
	"log/slog"
	"sort"
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

const (
	fohMinDistance    = 8
	fohMaxDistance    = 15
	hbMinDistance     = 6
	hbMaxDistance     = 12
	fohMaxAttacksLoop = 35              // Maximum attack attempts before resetting
	castingTimeout    = 3 * time.Second // Maximum time to wait for a cast to complete
)

type Foh struct {
	BaseCharacter
	lastCastTime time.Time
}

func (s Foh) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.Conviction, skill.HolyShield, skill.TomeOfTownPortal, skill.FistOfTheHeavens, skill.HolyBolt}
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

// waitForCastComplete waits until the character is no longer in casting animation
func (f *Foh) waitForCastComplete() bool {
	ctx := context.Get()
	startTime := time.Now()

	for time.Since(startTime) < castingTimeout {
		ctx.RefreshGameData()

		// Check if we're no longer casting and enough time has passed since last cast
		if ctx.Data.PlayerUnit.Mode != mode.CastingSkill &&
			time.Since(f.lastCastTime) > 150*time.Millisecond { //150 for Foh but if we make that generic it would need tuning maybe from skill desc
			return true
		}

		time.Sleep(16 * time.Millisecond) // Small sleep to avoid hammering CPU
	}

	return false
}
func (s Foh) KillMonsterSequence(monsterSelector func(d game.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist) error {
	ctx := context.Get()
	lastRefresh := time.Now()
	completedAttackLoops := 0
	var currentTargetID data.UnitID
	useHolyBolt := false

	fohOpts := []step.AttackOption{
		step.StationaryDistance(fohMinDistance, fohMaxDistance),
		step.EnsureAura(skill.Conviction),
	}
	hbOpts := []step.AttackOption{
		step.StationaryDistance(hbMinDistance, hbMaxDistance),
		step.EnsureAura(skill.Conviction),
	}

	// Initial target selection and analysis
	initialTargetAnalysis := func() (data.UnitID, bool, bool) {
		id, found := monsterSelector(*s.Data)
		if !found {
			return 0, false, false
		}

		// Count initial valid targets
		validTargets := 0
		monstersInRange := make([]data.Monster, 0)
		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			return 0, false, false
		}

		for _, m := range ctx.Data.Monsters.Enemies() {
			if ctx.Data.AreaData.IsInside(m.Position) {
				dist := ctx.PathFinder.DistanceFromMe(m.Position)
				if dist <= fohMaxDistance && dist >= fohMinDistance && m.Stats[stat.Life] > 0 {
					validTargets++
					monstersInRange = append(monstersInRange, m)
				}
			}
		}

		// Determine if we should use Holy Bolt
		// Only use Holy Bolt if it's a single target and it's immune to lightning
		shouldUseHB := validTargets == 1 && monster.IsImmune(stat.LightImmune)

		return id, true, shouldUseHB
	}

	for {
		// Refresh game data periodically
		if time.Since(lastRefresh) > time.Millisecond*100 {
			ctx.RefreshGameData()
			lastRefresh = time.Now()
		}

		ctx.PauseIfNotPriority()

		if completedAttackLoops >= fohMaxAttacksLoop {
			return nil
		}

		// If we don't have a current target, get one and analyze the situation
		if currentTargetID == 0 {
			var found bool
			currentTargetID, found, useHolyBolt = initialTargetAnalysis()
			if !found {
				return nil
			}
		}

		// Verify our target still exists and is alive
		monster, found := s.Data.Monsters.FindByID(currentTargetID)
		if !found || monster.Stats[stat.Life] <= 0 {
			currentTargetID = 0 // Reset target
			continue
		}

		if !s.preBattleChecks(currentTargetID, skipOnImmunities) {
			return nil
		}

		// Ensure Conviction is active
		if !s.Data.PlayerUnit.States.HasState(state.Conviction) {
			if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.Conviction); found {
				ctx.HID.PressKeyBinding(kb)
			}
		}

		// Cast appropriate skill
		if useHolyBolt {
			if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.HolyBolt); found {
				ctx.HID.PressKeyBinding(kb)
				if err := step.PrimaryAttack(currentTargetID, 1, true, hbOpts...); err == nil {
					if !s.waitForCastComplete() {
						continue
					}
					s.lastCastTime = time.Now()
					completedAttackLoops++
				}
			}
		} else {
			if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.FistOfTheHeavens); found {
				ctx.HID.PressKeyBinding(kb)
				if err := step.PrimaryAttack(currentTargetID, 1, true, fohOpts...); err == nil {
					if !s.waitForCastComplete() {
						continue
					}
					s.lastCastTime = time.Now()
					completedAttackLoops++
				}
			}
		}
	}
}

func (f *Foh) handleBoss(bossID data.UnitID, fohOpts, hbOpts []step.AttackOption, completedAttackLoops *int) error {
	ctx := context.Get()

	// Cast FoH
	if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.FistOfTheHeavens); found {
		ctx.HID.PressKeyBinding(kb)

		if err := step.PrimaryAttack(bossID, 1, true, fohOpts...); err == nil {
			// Wait for FoH cast to complete
			if !f.waitForCastComplete() {
				return fmt.Errorf("FoH cast timed out")
			}
			f.lastCastTime = time.Now()

			// Switch to Holy Bolt
			if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.HolyBolt); found {
				ctx.HID.PressKeyBinding(kb)

				// Cast 3 Holy Bolts
				for i := 0; i < 3; i++ {
					if err := step.PrimaryAttack(bossID, 1, true, hbOpts...); err == nil {
						if !f.waitForCastComplete() {
							return fmt.Errorf("Holy Bolt cast timed out")
						}
						f.lastCastTime = time.Now()
					}
				}

				(*completedAttackLoops)++
			}
		}
	}
	return nil
}
func (s Foh) KillBossSequence(monsterSelector func(d game.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist) error {
	ctx := context.Get()
	lastRefresh := time.Now()
	completedAttackLoops := 0
	fohOpts := []step.AttackOption{
		step.StationaryDistance(fohMinDistance, fohMaxDistance),
		step.EnsureAura(skill.Conviction),
	}
	hbOpts := []step.AttackOption{
		step.StationaryDistance(hbMinDistance, hbMaxDistance),
		step.EnsureAura(skill.Conviction),
	}

	for {
		if time.Since(lastRefresh) > time.Millisecond*100 {
			ctx.RefreshGameData()
			lastRefresh = time.Now()
		}
		ctx.PauseIfNotPriority()
		if completedAttackLoops >= fohMaxAttacksLoop {
			return nil
		}
		id, found := monsterSelector(*s.Data)
		if !found {
			return nil
		}
		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}
		monster, found := s.Data.Monsters.FindByID(id)
		if !found || monster.Stats[stat.Life] <= 0 {
			return nil
		}
		if !s.Data.PlayerUnit.States.HasState(state.Conviction) {
			if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.Conviction); found {
				ctx.HID.PressKeyBinding(kb)
			}
		}

		if err := s.handleBoss(monster.UnitID, fohOpts, hbOpts, &completedAttackLoops); err == nil {
			continue
		}
	}
}

func (s Foh) BuffSkills() []skill.ID {
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.HolyShield); found {
		return []skill.ID{skill.HolyShield}
	}
	return []skill.ID{}
}

func (s Foh) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s Foh) killBoss(npc npc.ID, t data.MonsterType) error {
	return s.KillBossSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found || m.Stats[stat.Life] <= 0 {
			return 0, false
		}
		return m.UnitID, true
	}, nil)
}

func (s Foh) KillCountess() error {
	return s.killBoss(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s Foh) KillAndariel() error {
	return s.killBoss(npc.Andariel, data.MonsterTypeUnique)
}

func (s Foh) KillSummoner() error {
	return s.killBoss(npc.Summoner, data.MonsterTypeUnique)
}

func (s Foh) KillDuriel() error {
	return s.killBoss(npc.Duriel, data.MonsterTypeUnique)
}

func (s Foh) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		var councilMembers []data.Monster
		for _, m := range d.Monsters {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				councilMembers = append(councilMembers, m)
			}
		}
		sort.Slice(councilMembers, func(i, j int) bool {
			return s.PathFinder.DistanceFromMe(councilMembers[i].Position) < s.PathFinder.DistanceFromMe(councilMembers[j].Position)
		})
		if len(councilMembers) > 0 {
			return councilMembers[0].UnitID, true
		}
		return 0, false
	}, nil)
}

func (s Foh) KillMephisto() error {
	return s.killBoss(npc.Mephisto, data.MonsterTypeUnique)
}

func (s Foh) KillIzual() error {
	return s.killBoss(npc.Izual, data.MonsterTypeUnique)
}

func (s Foh) KillDiablo() error {
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

		return s.killBoss(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (s Foh) KillPindle() error {
	return s.killBoss(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s Foh) KillNihlathak() error {
	return s.killBoss(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s Foh) KillBaal() error {
	return s.killBoss(npc.BaalCrab, data.MonsterTypeUnique)
}
