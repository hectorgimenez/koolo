package character

import (
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/mode"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	druMaxAttacksLoop   = 20              // Max number of attack loops before stopping
	druMinDistance      = 2               // Min distance to maintain from target
	druMaxDistance      = 8               // Max distance to maintain from target
	druidCastingTimeout = 3 * time.Second // Timeout for casting actions
)

type WindDruid struct {
	BaseCharacter           // Inherits common character functionality
	lastCastTime  time.Time // Tracks the last time a skill was cast
}

// Verify that required skills are bound to keys
func (s WindDruid) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.Hurricane, skill.OakSage, skill.CycloneArmor, skill.TomeOfTownPortal, skill.Tornado}
	missingKeybindings := make([]skill.ID, 0)

	for _, cskill := range requireKeybindings {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		s.Logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings // Returns list of unbound required skills
}

// Ensure casting animation finishes before proceeding
func (s WindDruid) waitForCastComplete() bool {
	ctx := context.Get()
	startTime := time.Now()

	for time.Since(startTime) < castingTimeout {
		ctx.RefreshGameData()

		if ctx.Data.PlayerUnit.Mode != mode.CastingSkill && // Check if not casting
			time.Since(s.lastCastTime) > 150*time.Millisecond { // Ensure enough time has passed
			return true
		}

		time.Sleep(16 * time.Millisecond) // Small delay to avoid busy-waiting
	}

	return false // Returns false if timeout is reached
}

// Handle the main combat loop for attacking monsters
func (s WindDruid) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool), // Function to select target monster
	skipOnImmunities []stat.Resist, // Resistances to skip if monster is immune
) error {
	ctx := context.Get()
	lastRefresh := time.Now()
	completedAttackLoops := 0
	var currentTargetID data.UnitID

	defer func() { // Ensures Tornado is set as active skill on exit
		if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.Tornado); found {
			ctx.HID.PressKeyBinding(kb)
		}
	}()

	attackOpts := []step.AttackOption{
		step.StationaryDistance(druMinDistance, druMaxDistance), // Maintains distance range
	}

	for {
		if time.Since(lastRefresh) > time.Millisecond*100 {
			ctx.RefreshGameData()
			lastRefresh = time.Now()
		}

		ctx.PauseIfNotPriority() // Pause if not the priority task

		if completedAttackLoops >= druMaxAttacksLoop {
			return nil // Exit if max attack loops reached
		}

		if currentTargetID == 0 { // Select a new target if none exists
			id, found := monsterSelector(*s.Data)
			if !found {
				return nil // Exit if no target found
			}
			currentTargetID = id
		}

		monster, found := s.Data.Monsters.FindByID(currentTargetID)
		if !found || monster.Stats[stat.Life] <= 0 { // Check if target is dead or missing
			currentTargetID = 0
			continue
		}

		if !s.preBattleChecks(currentTargetID, skipOnImmunities) { // Perform pre-combat checks
			return nil
		}

		s.RecastBuffs() // Refresh buffs before attacking

		if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.Tornado); found {
			ctx.HID.PressKeyBinding(kb) // Set Tornado as active skill
			if err := step.PrimaryAttack(currentTargetID, 1, true, attackOpts...); err == nil {
				if !s.waitForCastComplete() { // Wait for cast to complete
					continue
				}
				s.lastCastTime = time.Now() // Update last cast time
				completedAttackLoops++
			}
		} else {
			return fmt.Errorf("tornado skill not bound")
		}
	}
}

// Helper for killing a specific monster by NPC ID and type
func (s WindDruid) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}
		return m.UnitID, true
	}, nil)
}

// Reapplies active buffs if they’ve expired
func (s WindDruid) RecastBuffs() {
	ctx := context.Get()
	skills := []skill.ID{skill.Hurricane, skill.OakSage, skill.CycloneArmor}
	states := []state.State{state.Hurricane, state.Oaksage, state.Cyclonearmor}

	for i, druSkill := range skills {
		if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(druSkill); found {
			if !ctx.Data.PlayerUnit.States.HasState(states[i]) { // Check if buff is missing
				ctx.HID.PressKeyBinding(kb)             // Activate skill
				utils.Sleep(180)                        // Small delay
				s.HID.Click(game.RightButton, 640, 340) // Cast skill at center screen
				utils.Sleep(100)                        // Delay to ensure cast completes
			}
		}
	}
}

// Return a list of available buff skills
func (s WindDruid) BuffSkills() []skill.ID {
	buffs := make([]skill.ID, 0)
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.CycloneArmor); found {
		buffs = append(buffs, skill.CycloneArmor)
	}
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.Raven); found {
		buffs = append(buffs, skill.Raven, skill.Raven, skill.Raven, skill.Raven, skill.Raven) // Summon 5 ravens
	}
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.Hurricane); found {
		buffs = append(buffs, skill.Hurricane)
	}
	return buffs
}

// Dynamically determines pre-combat buffs and summons
func (s WindDruid) PreCTABuffSkills() []skill.ID {
	needsBear := true
	wolves := 5
	direWolves := 3
	needsOak := true
	needsCreeper := true

	for _, monster := range s.Data.Monsters { // Check existing pets
		if monster.IsPet() {
			switch monster.Name {
			case npc.DruBear:
				needsBear = false
			case npc.DruFenris:
				direWolves--
			case npc.DruSpiritWolf:
				wolves--
			case npc.OakSage:
				needsOak = false
			case npc.DruCycleOfLife, npc.VineCreature, npc.DruPlaguePoppy:
				needsCreeper = false
			}
		}
	}

	skills := make([]skill.ID, 0)
	if s.Data.PlayerUnit.States.HasState(state.Oaksage) {
		needsOak = false
	}

	// Add summoning skills based on need and key bindings
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.SummonSpiritWolf); found {
		for i := 0; i < wolves; i++ {
			skills = append(skills, skill.SummonSpiritWolf)
		}
	}
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.SummonDireWolf); found {
		for i := 0; i < direWolves; i++ {
			skills = append(skills, skill.SummonDireWolf)
		}
	}
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.SummonGrizzly); found && needsBear {
		skills = append(skills, skill.SummonGrizzly)
	}
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.OakSage); found && needsOak {
		skills = append(skills, skill.OakSage)
	}
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.SolarCreeper); found && needsCreeper {
		skills = append(skills, skill.SolarCreeper)
	}
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.CarrionVine); found && needsCreeper {
		skills = append(skills, skill.CarrionVine)
	}
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.PoisonCreeper); found && needsCreeper {
		skills = append(skills, skill.PoisonCreeper)
	}

	return skills
}

func (s WindDruid) KillCountess() error {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s WindDruid) KillAndariel() error {
	return s.killMonster(npc.Andariel, data.MonsterTypeUnique)
}

func (s WindDruid) KillSummoner() error {
	return s.killMonster(npc.Summoner, data.MonsterTypeUnique)
}

func (s WindDruid) KillDuriel() error {
	return s.killMonster(npc.Duriel, data.MonsterTypeUnique)
}

// Targets multiple council members, sorted by distance
func (s WindDruid) KillCouncil() error {
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

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (s WindDruid) KillMephisto() error {
	return s.killMonster(npc.Mephisto, data.MonsterTypeUnique)
}

func (s WindDruid) KillIzual() error {
	return s.killMonster(npc.Izual, data.MonsterTypeUnique)
}

// KillDiablo includes a timeout and detection logic
func (s WindDruid) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			s.Logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := s.Data.Monsters.FindOne(npc.Diablo, data.MonsterTypeUnique)
		if !found || diablo.Stats[stat.Life] <= 0 { // Check if Diablo is dead or missing
			// Diablo is dead
			if diabloFound {
				return nil
			}
			// Keep waiting..
			time.Sleep(200 * time.Millisecond)
			continue
		}

		diabloFound = true
		s.Logger.Info("Diablo detected, attacking")
		return s.killMonster(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (s WindDruid) KillPindle() error {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s WindDruid) KillNihlathak() error {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s WindDruid) KillBaal() error {
	return s.killMonster(npc.BaalCrab, data.MonsterTypeUnique)
}
