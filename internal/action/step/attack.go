package step

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/utils"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	attackCycleDuration    = 120 * time.Millisecond
	healthCheckCooldown    = 100 * time.Millisecond
	positionStuckThreshold = 3 * time.Second
	stateCleanupInterval   = 5 * time.Minute
)

var (
	statesMutex   sync.RWMutex
	monsterStates = make(map[data.UnitID]*attackState)
)

// Contains all configuration for an attack sequence
type attackSettings struct {
	primaryAttack    bool          // Whether this is a primary (left click) attack
	skill            skill.ID      // Skill ID for secondary attacks
	followEnemy      bool          // Whether to follow the enemy while attacking
	minDistance      int           // Minimum attack range
	maxDistance      int           // Maximum attack range
	aura             skill.ID      // Aura to maintain during attack
	target           data.UnitID   // Specific target's unit ID (0 for AOE)
	shouldStandStill bool          // Whether to stand still while attacking
	numOfAttacks     int           // Number of attacks to perform
	timeout          time.Duration // Timeout for the attack sequence
	isBurstCastSkill bool          // Whether this is a channeled/burst skill like Nova
}

// AttackOption defines a function type for configuring attack settings
type AttackOption func(*attackSettings)

type attackState struct {
	lastHealth             int
	lastHealthCheckTime    time.Time
	failedAttemptStartTime time.Time
	position               data.Position
}

// Distance configures attack to follow enemy within specified range
func Distance(min, max int) AttackOption {
	return func(s *attackSettings) {
		s.followEnemy = true
		s.minDistance = min
		s.maxDistance = max
	}
}

// RangedDistance configures attack for ranged combat without following
func RangedDistance(min, max int) AttackOption {
	return func(s *attackSettings) {
		s.followEnemy = false // Don't follow enemies for ranged attacks
		s.minDistance = min
		s.maxDistance = max
	}
}

// StationaryDistance configures attack to remain stationary (like FoH)
func StationaryDistance(min, max int) AttackOption {
	return func(s *attackSettings) {
		s.followEnemy = false
		s.minDistance = min
		s.maxDistance = max
		s.shouldStandStill = true
	}
}

// EnsureAura ensures specified aura is active during attack
func EnsureAura(aura skill.ID) AttackOption {
	return func(s *attackSettings) {
		s.aura = aura
	}
}

// PrimaryAttack initiates a primary (left-click) attack sequence
func PrimaryAttack(target data.UnitID, attacks int, standStill bool, opts ...AttackOption) error {
	ctx := context.Get()

	// Special handling for Berserker characters
	if berserker, ok := ctx.Char.(interface{ PerformBerserkAttack(data.UnitID) }); ok {
		for i := 0; i < attacks; i++ {
			berserker.PerformBerserkAttack(target)
		}
		return nil
	}

	settings := attackSettings{
		target:           target,
		numOfAttacks:     attacks,
		shouldStandStill: standStill,
		primaryAttack:    true,
	}
	for _, o := range opts {
		o(&settings)
	}

	return executeAttackSequence(settings)
}

// SecondaryAttack initiates a secondary (right-click) attack sequence with a specific skill
func SecondaryAttack(skill skill.ID, target data.UnitID, attacks int, opts ...AttackOption) error {
	settings := attackSettings{
		target:           target,
		numOfAttacks:     attacks,
		skill:            skill,
		isBurstCastSkill: skill == 48, // Nova skill ID
	}
	for _, o := range opts {
		o(&settings)
	}

	if settings.isBurstCastSkill {
		settings.timeout = 30 * time.Second
		return executeBurstAttack(settings)
	}
	return executeAttackSequence(settings)
}

// Helper function to validate if a monster should be targetable
func isValidEnemy(m data.Monster, ctx *context.Status) bool {
	// Special case: Always allow Vizier seal boss even if off grid
	isVizier := m.Type == data.MonsterTypeSuperUnique && m.Name == npc.StormCaster
	if isVizier {
		return m.Stats[stat.Life] > 0
	}

	// Skip monsters in invalid positions
	if !ctx.Data.AreaData.IsWalkable(m.Position) {
		return false
	}

	// Skip dead monsters
	if m.Stats[stat.Life] <= 0 {
		return false
	}

	return true
}

// Cleanup function to ensure proper state on exit
func keyCleanup(ctx *context.Status) {
	ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
}

// executeAttackSequence handles the core attack logic for both primary and secondary attacks
func executeAttackSequence(settings attackSettings) error {
	ctx := context.Get()
	ctx.SetLastStep("Attack")
	defer keyCleanup(ctx)

	remainingAttacks := settings.numOfAttacks
	var lastAttack time.Time

	for remainingAttacks > 0 {
		ctx.PauseIfNotPriority()

		// Check target validity
		target, valid := ctx.Data.Monsters.FindByID(settings.target)
		if !valid || !isValidEnemy(target, ctx) {
			return nil
		}

		// Position validation
		distance := ctx.PathFinder.DistanceFromMe(target.Position)
		if !lastAttack.IsZero() && !settings.followEnemy && distance > settings.maxDistance {
			return nil // Enemy out of range with no following
		}

		// Damage and positioning checks
		_, state := checkMonsterDamage(target)
		needsReposition := !state.failedAttemptStartTime.IsZero() &&
			time.Since(state.failedAttemptStartTime) > positionStuckThreshold

		// Maintain optimal attack position
		if err := ensureEnemyIsInRange(target, settings.maxDistance, settings.minDistance, needsReposition); err != nil {
			return fmt.Errorf("positioning failure: %w", err)
		}

		// Aura management
		if settings.aura != 0 && lastAttack.IsZero() {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.aura))
		}

		// Attack timing control
		if time.Since(lastAttack) > ctx.Data.PlayerCastDuration()-attackCycleDuration {
			performAttack(ctx, settings, target.Position.X, target.Position.Y)
			lastAttack = time.Now()
			remainingAttacks--
		}
	}
	return nil
}

// executeBurstAttack handles channeled skills requiring continuous execution
func executeBurstAttack(settings attackSettings) error {
	ctx := context.Get()
	ctx.SetLastStep("BurstAttack")
	defer keyCleanup(ctx)

	startTime := time.Now()
	for time.Since(startTime) < settings.timeout {
		ctx.PauseIfNotPriority()

		// Find valid target in range
		var target data.Monster
		for _, m := range ctx.Data.Monsters.Enemies() {
			if isValidEnemy(m, ctx) && ctx.PathFinder.DistanceFromMe(m.Position) <= settings.maxDistance {
				target = m
				break
			}
		}

		if target.UnitID == 0 {
			return nil // No valid targets
		}

		// Positioning checks
		_, state := checkMonsterDamage(target)
		needsReposition := !state.failedAttemptStartTime.IsZero() &&
			time.Since(state.failedAttemptStartTime) > positionStuckThreshold

		// Line of sight validation
		if !ctx.PathFinder.LineOfSight(ctx.Data.PlayerUnit.Position, target.Position) || needsReposition {
			if err := ensureEnemyIsInRange(target, settings.maxDistance, settings.minDistance, needsReposition); err != nil {
				return fmt.Errorf("reposition failure: %w", err)
			}
			continue
		}

		performAttack(ctx, settings, target.Position.X, target.Position.Y)
	}
	return nil
}

// performAttack executes the actual game input for attacks
func performAttack(ctx *context.Status, settings attackSettings, x, y int) {
	targetPos := data.Position{X: x, Y: y}
	if !ctx.PathFinder.LineOfSight(ctx.Data.PlayerUnit.Position, targetPos) {
		return // Skip attacks without line of sight
	}

	// Skill selection for secondary attacks
	if settings.skill != 0 && ctx.Data.PlayerUnit.RightSkill != settings.skill {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.skill))
		time.Sleep(10 * time.Millisecond)
	}

	// Stand still handling
	if settings.shouldStandStill {
		ctx.HID.KeyDown(ctx.Data.KeyBindings.StandStill)
		defer ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
	}

	// Convert coordinates and send click
	screenX, screenY := ctx.PathFinder.GameCoordsToScreenCords(x, y)
	button := game.LeftButton
	if !settings.primaryAttack {
		button = game.RightButton
	}
	ctx.HID.Click(button, screenX, screenY)
}

// ensureEnemyIsInRange handles character positioning relative to targets
func ensureEnemyIsInRange(monster data.Monster, maxDist, minDist int, needsReposition bool) error {
	ctx := context.Get()
	currentPos := ctx.Data.PlayerUnit.Position
	distance := ctx.PathFinder.DistanceFromMe(monster.Position)

	// Early exit if already in position
	if ctx.PathFinder.LineOfSight(currentPos, monster.Position) && distance <= maxDist && !needsReposition {
		return nil
	}

	// Force reposition if stuck
	if needsReposition {
		ctx.Logger.Info("Repositioning due to attack failure")
		dest := ctx.PathFinder.BeyondPosition(currentPos, monster.Position, 4)
		return MoveTo(dest)
	}

	// Melee character handling
	if maxDist <= 3 {
		return MoveTo(monster.Position)
	}

	// Pathfinding implementation
	path, _, found := ctx.PathFinder.GetPath(monster.Position)
	if !found {
		return errors.New("unreachable target")
	}

	// Find optimal attack position along path
	for _, pos := range path {
		monsterDistance := utils.DistanceFromPoint(ctx.Data.AreaData.RelativePosition(monster.Position), pos)
		if monsterDistance > maxDist || monsterDistance < minDist {
			continue
		}

		dest := data.Position{
			X: pos.X + ctx.Data.AreaData.OffsetX,
			Y: pos.Y + ctx.Data.AreaData.OffsetY,
		}

		// Handle short-distance overshooting
		if ctx.PathFinder.DistanceFromMe(dest) <= DistanceToFinishMoving {
			dest = ctx.PathFinder.BeyondPosition(currentPos, dest, 9)
		}

		if ctx.PathFinder.LineOfSight(dest, monster.Position) {
			return MoveTo(dest)
		}
	}
	return errors.New("no valid attack position found")
}

// checkMonsterDamage tracks target health changes for effectiveness validation
func checkMonsterDamage(monster data.Monster) (bool, *attackState) {
	statesMutex.Lock()
	defer statesMutex.Unlock()

	state, exists := monsterStates[monster.UnitID]
	if !exists {
		state = &attackState{
			lastHealth:          monster.Stats[stat.Life],
			lastHealthCheckTime: time.Now(),
			position:            monster.Position,
		}
		monsterStates[monster.UnitID] = state
	}

	// Health change detection
	if time.Since(state.lastHealthCheckTime) > healthCheckCooldown {
		currentHealth := monster.Stats[stat.Life]
		now := time.Now()

		if currentHealth < state.lastHealth {
			state.failedAttemptStartTime = time.Time{}
		} else if state.failedAttemptStartTime.IsZero() && monster.Position == state.position {
			state.failedAttemptStartTime = now
		}

		state.lastHealth = currentHealth
		state.lastHealthCheckTime = now
		state.position = monster.Position

		// Periodic state cleanup
		if len(monsterStates) > 100 {
			now := time.Now()
			for id, s := range monsterStates {
				if now.Sub(s.lastHealthCheckTime) > stateCleanupInterval {
					delete(monsterStates, id)
				}
			}
		}
	}
	return state.lastHealth < monster.Stats[stat.Life], state
}
