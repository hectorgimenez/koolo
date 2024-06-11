package run

import (
	"log/slog"
	"slices"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/pather"
)

var diabloSpawnPosition = data.Position{
	X: 7800,
	Y: 5286,
}

var chaosSanctuaryEntrancePosition = data.Position{
	X: 7790,
	Y: 5544,
}

var entranceToStar = []data.Position{
	{X: 7791, Y: 5491},
	{X: 7768, Y: 5459},
	{X: 7775, Y: 5424},
	{X: 7817, Y: 5458},
	{X: 7777, Y: 5408},
	{X: 7769, Y: 5379},
	{X: 7777, Y: 5357},
	{X: 7809, Y: 5359},
	{X: 7805, Y: 5330},
	{X: 7780, Y: 5317},
}

var starToViz = []data.Position{
	{X: 7760, Y: 5295},
	{X: 7744, Y: 5295},
	{X: 7710, Y: 5290},
	{X: 7675, Y: 5290},
	{X: 7665, Y: 5315},
	{X: 7665, Y: 5275},
}

var starToSeis = []data.Position{
	{X: 7790, Y: 5255},
	{X: 7790, Y: 5230},
	{X: 7770, Y: 5205},
	{X: 7813, Y: 5190},
	{X: 7813, Y: 5158},
	{X: 7790, Y: 5155},
}

var starToInf = []data.Position{
	{X: 7825, Y: 5290},
	{X: 7845, Y: 5290},
	{X: 7870, Y: 5277},
	{X: 7933, Y: 5316},
}

type Diablo struct {
	baseRun
	bm health.BeltManager
}

func (a Diablo) Name() string {
	return string(config.DiabloRun)
}

func (a Diablo) BuildActions() (actions []action.Action) {
	actions = append(actions,
		// Moving to starting point (RiverOfFlame)
		a.builder.WayPoint(area.RiverOfFlame),
		a.builder.MoveToCoords(chaosSanctuaryEntrancePosition),
	)

	// Let's move to a safe area and open the portal in companion mode
	if a.CharacterCfg.Companion.Enabled && a.CharacterCfg.Companion.Leader {
		actions = append(actions,
			a.builder.MoveToCoords(diabloSpawnPosition),
			a.builder.OpenTPIfLeader(),
			a.builder.ClearAreaAroundPlayer(50, data.MonsterAnyFilter()),
		)
	}

	if a.Container.CharacterCfg.Game.Diablo.ClearArea {
		monsterFilter := data.MonsterAnyFilter()
		if a.Container.CharacterCfg.Game.Diablo.OnlyElites {
			monsterFilter = data.MonsterEliteFilter()
		}

		actions = slices.Concat(actions,
			a.generateClearActions(entranceToStar, monsterFilter),
			a.generateClearActions(starToViz, monsterFilter),
			a.generateClearActions(starToSeis, monsterFilter),
			a.generateClearActions(starToInf, monsterFilter),
		)
	} else {
		actions = append(actions,
			// Travel to diablo spawn location
			a.builder.MoveToCoords(diabloSpawnPosition),
		)
	}

	seals := []object.Name{object.DiabloSeal4, object.DiabloSeal5, object.DiabloSeal3, object.DiabloSeal2, object.DiabloSeal1}

	// Move across all the seals, try to find the most clear spot around them, kill monsters and activate the seal.
	for i, s := range seals {
		seal := s
		sealNumber := i

		actions = append(actions, action.NewChain(func(d game.Data) []action.Action {
			_, isLevelingChar := a.char.(action.LevelingCharacter)
			if isLevelingChar && (a.bm.ShouldBuyPotions(d) || (a.CharacterCfg.Character.UseMerc && d.MercHPPercent() <= 0)) {
				a.logger.Debug("Let's go back town to buy more pots", slog.Int("seal", sealNumber+1))
				return a.builder.InRunReturnTownRoutine()
			}

			return nil
		}))

		actions = append(actions, a.builder.MoveTo(func(d game.Data) (data.Position, bool) {
			a.logger.Debug("Moving to next seal", slog.Int("seal", sealNumber+1))
			if obj, found := d.Objects.FindOne(seal); found {
				if d := pather.DistanceFromMe(d, obj.Position); d < 7 {
					a.logger.Debug("We are close enough to the seal", slog.Int("seal", sealNumber+1))
					return data.Position{}, false
				}

				a.logger.Debug("Seal found, start teleporting", slog.Int("seal", sealNumber+1))

				return obj.Position, true
			}
			a.logger.Debug("Seal NOT found", slog.Int("seal", sealNumber+1))

			return data.Position{}, false
		}, step.StopAtDistance(7)))

		// Try to calculate based on a square boundary around the seal which corner is safer, then tele there
		//actions = append(actions, action.NewStepChain(func(d game.Data) []step.Step {
		//	if obj, found := d.Objects.FindOne(seal); found {
		//		pos := a.getLessConcurredCornerAroundSeal(d, obj.Position)
		//		return []step.Step{step.MoveTo(pos)}
		//	}
		//	return []step.Step{}
		//}))

		// Kill all the monsters close to the seal and item pickup
		actions = append(actions,
			a.builder.ClearAreaAroundPlayer(13, data.MonsterAnyFilter()),
			a.builder.ItemPickup(false, 40),
		)

		// Activate the seal, buff only before opening the first seal
		actions = append(actions,
			a.builder.ClearAreaAroundPlayer(15, data.MonsterAnyFilter()),
			action.NewChain(func(d game.Data) []action.Action {
				if i == 0 {
					return []action.Action{
						a.builder.Buff(),
					}
				}
				return []action.Action{}
			}),
		)

		actions = append(actions, action.NewChain(func(d game.Data) []action.Action {
			obj, _ := d.Objects.FindOne(seal)
			// Bugged seal, we need to move to a specific position to activate it
			if obj.Position.X == 7773 && obj.Position.Y == 5155 {
				return []action.Action{
					a.builder.MoveToCoords(data.Position{
						X: 7768,
						Y: 5160,
					}),
					action.NewStepChain(func(d game.Data) []step.Step {
						return []step.Step{step.InteractObjectByID(obj.ID, func(d game.Data) bool {
							if obj, found := d.Objects.FindOne(seal); found {
								if !obj.Selectable {
									a.logger.Debug("Seal activated, waiting for elite group to spawn", slog.Int("seal", sealNumber+1))
								}
								return !obj.Selectable
							}
							return false
						})}
					}),
				}
			}

			return []action.Action{a.builder.InteractObject(seal, func(d game.Data) bool {
				if obj, found := d.Objects.FindOne(seal); found {
					if !obj.Selectable {
						a.logger.Debug("Seal activated, waiting for elite group to spawn", slog.Int("seal", sealNumber+1))
					}
					return !obj.Selectable
				}
				return false
			})}
		}))

		// Only if we are not in the first seal
		if sealNumber != 0 {
			if sealNumber == 2 {
				actions = append(actions, a.builder.MoveToCoords(data.Position{
					X: 7773,
					Y: 5195,
				}))
			}

			// Now wait & try to kill the Elite packs (maybe are already dead, killed during previous action)
			startTime := time.Time{}
			actions = append(actions, action.NewStepChain(func(d game.Data) []step.Step {
				if startTime.IsZero() {
					startTime = time.Now()
				}
				for _, m := range d.Monsters.Enemies(data.MonsterEliteFilter()) {
					if a.builder.IsMonsterSealElite(m) {
						a.logger.Debug("Seal defender found!")
						return nil
					}
				}

				if time.Since(startTime) < time.Second*5 {
					return []step.Step{step.Wait(time.Millisecond * 100)}
				}

				return nil
			}, action.RepeatUntilNoSteps()))

			actions = append(actions, a.char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
				for _, m := range d.Monsters.Enemies(data.MonsterEliteFilter()) {
					if a.builder.IsMonsterSealElite(m) {
						_, _, found := a.PathFinder.GetPath(d, m.Position)
						return m.UnitID, found
					}
				}

				return 0, false
			}, nil))
		}

		actions = append(actions, a.builder.ItemPickup(false, 40))
	}

	_, isLevelingChar := a.char.(action.LevelingCharacter)

	// For leveling we always want to kill Diablo
	if isLevelingChar || a.Container.CharacterCfg.Game.Diablo.KillDiablo {
		// Go back to town to buy potions if needed
		actions = append(actions, action.NewChain(func(d game.Data) []action.Action {
			if isLevelingChar && (a.bm.ShouldBuyPotions(d) || (a.CharacterCfg.Character.UseMerc && d.MercHPPercent() <= 0)) {
				return a.builder.InRunReturnTownRoutine()
			}

			return nil
		}))

		actions = append(actions,
			a.builder.Buff(),
			a.builder.MoveToCoords(diabloSpawnPosition),
			a.char.KillDiablo(),
		)
	}

	return
}

func (a Diablo) getLessConcurredCornerAroundSeal(d game.Data, sealPosition data.Position) data.Position {
	corners := [4]data.Position{
		{
			X: sealPosition.X + 7,
			Y: sealPosition.Y + 7,
		},
		{
			X: sealPosition.X - 7,
			Y: sealPosition.Y + 7,
		},
		{
			X: sealPosition.X - 7,
			Y: sealPosition.Y - 7,
		},
		{
			X: sealPosition.X + 7,
			Y: sealPosition.Y - 7,
		},
	}

	bestCorner := 0
	bestCornerDistance := 0
	for i, c := range corners {
		averageDistance := 0
		monstersFound := 0
		for _, m := range d.Monsters.Enemies() {
			distance := pather.DistanceFromPoint(c, m.Position)
			// Ignore enemies not close to the seal
			if distance < 5 {
				monstersFound++
				averageDistance += pather.DistanceFromPoint(c, m.Position)
			}
		}
		if averageDistance > bestCornerDistance {
			bestCorner = i
			bestCornerDistance = averageDistance
		}
		// No monsters found here, don't need to keep checking
		if monstersFound == 0 {
			a.logger.Debug("Moving to corner", slog.Int("corner", i), slog.Int("monsters", monstersFound))
			return corners[i]
		}
		a.logger.Debug("Corner", slog.Int("corner", i), slog.Int("monsters", monstersFound), slog.Int("distance", averageDistance))
	}

	a.logger.Debug("Moving to corner", slog.Int("corner", bestCorner), slog.Int("monsters", bestCornerDistance))

	return corners[bestCorner]
}

func (a Diablo) generateClearActions(positions []data.Position, filter data.MonsterFilter) []action.Action {
	var actions []action.Action
	var maxPosDiff = 20

	for _, pos := range positions {
		actions = append(actions,
			action.NewChain(func(d game.Data) []action.Action {
				multiplier := 1

				if pather.IsWalkable(pos, d.AreaOrigin, d.CollisionGrid) {
					return []action.Action{a.builder.MoveToCoords(pos)}
				}

				for _ = range 2 {
					for i := 1; i < maxPosDiff; i++ {
						// Adjusting both X and Y gave fewer errors in testing
						newPos := data.Position{X: pos.X + (i * multiplier), Y: pos.Y + (i * multiplier)}

						if pather.IsWalkable(newPos, d.AreaOrigin, d.CollisionGrid) {
							return []action.Action{a.builder.MoveToCoords(newPos)}
						}

					}
					// Switch from + to -
					multiplier *= -1
				}

				// Let it fail then
				return []action.Action{a.builder.MoveToCoords(pos)}
			}),
			// Skip storm casters for now completely while clearing non-seals
			a.builder.ClearAreaAroundPlayer(35, func(m data.Monsters) []data.Monster {
				var monsters []data.Monster

				monsters = filter(m)
				monsters = skipStormCasterFilter(monsters)

				return monsters
			}),
			a.builder.ItemPickup(false, 35),
		)
	}

	return actions
}

func skipStormCasterFilter(monsters data.Monsters) []data.Monster {
	var stormCasterIds = []npc.ID{npc.StormCaster, npc.StormCaster2}
	var filteredMonsters []data.Monster

	for _, m := range monsters {
		if !slices.Contains(stormCasterIds, m.Name) {
			filteredMonsters = append(filteredMonsters, m)
		}
	}

	return filteredMonsters
}
