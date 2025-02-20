package character

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/mode"
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

const (
	fohMinDistance    = 8
	fohMaxDistance    = 15
	hbMinDistance     = 6
	hbMaxDistance     = 12
	fohMaxAttacksLoop = 42              // Maximum attack attempts before resetting
	castingTimeout    = 3 * time.Second // Maximum time to wait for a cast to complete
)

type Foh struct {
	BaseCharacter
	lastCastTime time.Time
}

func (f Foh) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.Conviction, skill.HolyShield, skill.TomeOfTownPortal, skill.FistOfTheHeavens, skill.HolyBolt}
	missingKeybindings := make([]skill.ID, 0)

	for _, cskill := range requireKeybindings {
		if _, found := f.Data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		f.Logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

// waitForCastComplete waits until the character is no longer in casting animation
func (f Foh) waitForCastComplete() bool {
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
func (f Foh) KillMonsterSequence(monsterSelector func(d game.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist) error {
	ctx := context.Get()
	lastRefresh := time.Now()
	completedAttackLoops := 0
	var currentTargetID data.UnitID
	useHolyBolt := false

	// Ensure we always return to FoH when done
	defer func() {
		if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.FistOfTheHeavens); found {
			ctx.HID.PressKeyBinding(kb)
		}
	}()

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
		id, found := monsterSelector(*f.Data)
		if !found {
			return 0, false, false
		}

		// Count initial valid targets
		validTargets := 0
		//monstersInRange := make([]data.Monster, 0)
		monster, found := f.Data.Monsters.FindByID(id)
		if !found {
			return 0, false, false
		}

		for _, m := range ctx.Data.Monsters.Enemies() {
			if ctx.Data.AreaData.IsInside(m.Position) {
				dist := ctx.PathFinder.DistanceFromMe(m.Position)
				if dist <= fohMaxDistance && dist >= fohMinDistance && m.Stats[stat.Life] > 0 {
					validTargets++
					//monstersInRange = append(monstersInRange, m)
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
		monster, found := f.Data.Monsters.FindByID(currentTargetID)
		if !found || monster.Stats[stat.Life] <= 0 {
			currentTargetID = 0 // Reset target
			continue
		}

		if !f.preBattleChecks(currentTargetID, skipOnImmunities) {
			return nil
		}

		// Ensure Conviction is active
		if !f.Data.PlayerUnit.States.HasState(state.Conviction) {
			if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.Conviction); found {
				ctx.HID.PressKeyBinding(kb)
			}
		}

		// Cast appropriate skill
		if useHolyBolt || completedAttackLoops%6 == 0 {
			if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.HolyBolt); found {
				ctx.HID.PressKeyBinding(kb)
				if err := step.PrimaryAttack(currentTargetID, 1, true, hbOpts...); err == nil {
					if !f.waitForCastComplete() {
						continue
					}
					f.lastCastTime = time.Now()
					completedAttackLoops++
				}
			}
		} else {
			if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.FistOfTheHeavens); found {
				ctx.HID.PressKeyBinding(kb)
				if err := step.PrimaryAttack(currentTargetID, 1, true, fohOpts...); err == nil {
					if !f.waitForCastComplete() {
						continue
					}
					f.lastCastTime = time.Now()
					completedAttackLoops++
				}
			}
		}
	}
}
func (f Foh) handleBoss(bossID data.UnitID, fohOpts, hbOpts []step.AttackOption, completedAttackLoops *int) error {
	ctx := context.Get()

	// Cast FoH
	if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.FistOfTheHeavens); found {
		ctx.HID.PressKeyBinding(kb)

		if err := step.PrimaryAttack(bossID, 1, true, fohOpts...); err == nil {
			// Wait for FoH cast to complete
			if !f.waitForCastComplete() {
				return fmt.Errorf("foh cast timed out")
			}
			f.lastCastTime = time.Now()

			// Switch to Holy Bolt
			if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.HolyBolt); found {
				ctx.HID.PressKeyBinding(kb)

				// Cast 3 Holy Bolts
				for i := 0; i < 3; i++ {
					if err := step.PrimaryAttack(bossID, 1, true, hbOpts...); err == nil {
						if !f.waitForCastComplete() {
							return fmt.Errorf("holy Bolt cast timed out")
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
func (f Foh) KillBossSequence(monsterSelector func(d game.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist) error {
	ctx := context.Get()
	lastRefresh := time.Now()
	completedAttackLoops := 0

	// Ensure we always return to FoH when done
	defer func() {
		if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.FistOfTheHeavens); found {
			ctx.HID.PressKeyBinding(kb)
		}
	}()

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
		id, found := monsterSelector(*f.Data)
		if !found {
			return nil
		}
		if !f.preBattleChecks(id, skipOnImmunities) {
			return nil
		}
		monster, found := f.Data.Monsters.FindByID(id)
		if !found || monster.Stats[stat.Life] <= 0 {
			return nil
		}
		if !f.Data.PlayerUnit.States.HasState(state.Conviction) {
			if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.Conviction); found {
				ctx.HID.PressKeyBinding(kb)
			}
		}

		if err := f.handleBoss(monster.UnitID, fohOpts, hbOpts, &completedAttackLoops); err == nil {
			continue
		}
	}
}

func (f Foh) BuffSkills() []skill.ID {
	if _, found := f.Data.KeyBindings.KeyBindingForSkill(skill.HolyShield); found {
		return []skill.ID{skill.HolyShield}
	}
	return make([]skill.ID, 0)
}

func (f Foh) PreCTABuffSkills() []skill.ID {
	return make([]skill.ID, 0)
}

func (f Foh) killBoss(npc npc.ID, t data.MonsterType) error {
	return f.KillBossSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found || m.Stats[stat.Life] <= 0 {
			return 0, false
		}
		return m.UnitID, true
	}, nil)
}

func (f Foh) KillCountess() error {
	return f.killBoss(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (f Foh) KillAndariel() error {
	return f.killBoss(npc.Andariel, data.MonsterTypeUnique)
}

func (f Foh) KillSummoner() error {
	return f.killBoss(npc.Summoner, data.MonsterTypeUnique)
}

func (f Foh) KillDuriel() error {
	return f.killBoss(npc.Duriel, data.MonsterTypeUnique)
}

func (f Foh) KillCouncil() error {
	// Disable item pickup while killing council members
	context.Get().DisableItemPickup()
	defer context.Get().EnableItemPickup()

	err := f.killAllCouncilMembers()
	if err != nil {
		return err
	}

	// Wait a moment for items to settle
	time.Sleep(300 * time.Millisecond)

	// Re-enable item pickup and do a final pickup pass
	err = action.ItemPickup(40)
	if err != nil {
		f.Logger.Warn("Error during final item pickup after council", "error", err)
	}

	return nil
}
func (f Foh) killAllCouncilMembers() error {
	for {
		if !f.anyCouncilMemberAlive() {
			return nil
		}

		err := f.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
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

func (f Foh) anyCouncilMemberAlive() bool {
	for _, m := range f.Data.Monsters.Enemies() {
		if (m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3) && m.Stats[stat.Life] > 0 {
			return true
		}

	}
	return false
}

func (f Foh) KillMephisto() error {
	return f.killBoss(npc.Mephisto, data.MonsterTypeUnique)
}

func (f Foh) KillIzual() error {
	return f.killBoss(npc.Izual, data.MonsterTypeUnique)
}

func (f Foh) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			f.Logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := f.Data.Monsters.FindOne(npc.Diablo, data.MonsterTypeUnique)
		if !found || diablo.Stats[stat.Life] <= 0 {
			if diabloFound {
				return nil
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		diabloFound = true
		f.Logger.Info("Diablo detected, attacking")

		return f.killBoss(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (f Foh) KillPindle() error {
	return f.killBoss(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (f Foh) KillNihlathak() error {
	return f.killBoss(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (f Foh) KillBaal() error {
	return f.killBoss(npc.BaalCrab, data.MonsterTypeUnique)
}
