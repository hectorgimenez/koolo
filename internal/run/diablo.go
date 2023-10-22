package run

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/pather"
	"go.uber.org/zap"
)

var diabloSpawnPosition = data.Position{
	X: 7800,
	Y: 5286,
}

var chaosSanctuaryEntrancePosition = data.Position{
	X: 7796,
	Y: 5561,
}

type Diablo struct {
	baseRun
	bm health.BeltManager
}

func (a Diablo) Name() string {
	return "Diablo"
}

func (a Diablo) BuildActions() (actions []action.Action) {
	actions = append(actions,
		// Moving to starting point (RiverOfFlame)
		a.builder.WayPoint(area.RiverOfFlame),
		// Travel to diablo spawn location
		a.builder.MoveToCoords(chaosSanctuaryEntrancePosition),
		a.builder.MoveToCoords(diabloSpawnPosition),
	)

	seals := []object.Name{object.DiabloSeal4, object.DiabloSeal5, object.DiabloSeal3, object.DiabloSeal2, object.DiabloSeal1}

	// Move across all the seals, try to find the most clear spot around them, kill monsters and activate the seal.
	for i, s := range seals {
		seal := s
		sealNumber := i

		actions = append(actions, action.NewChain(func(d data.Data) []action.Action {
			_, isLevelingChar := a.char.(action.LevelingCharacter)
			if isLevelingChar && (a.bm.ShouldBuyPotions(d) || (config.Config.Character.UseMerc && d.MercHPPercent() <= 0)) {
				a.logger.Debug("Let's go back town to buy more pots", zap.Int("seal", sealNumber+1))
				return a.builder.InRunReturnTownRoutine()
			}

			return nil
		}))

		actions = append(actions, a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
			a.logger.Debug("Moving to next seal", zap.Int("seal", sealNumber+1))
			if obj, found := d.Objects.FindOne(seal); found {
				if d := pather.DistanceFromMe(d, obj.Position); d < 7 {
					a.logger.Debug("We are close enough to the seal", zap.Int("seal", sealNumber+1))
					return data.Position{}, false
				}

				a.logger.Debug("Seal found, start teleporting", zap.Int("seal", sealNumber+1))

				return obj.Position, true
			}
			a.logger.Debug("Seal NOT found", zap.Int("seal", sealNumber+1))

			return data.Position{}, false
		}, step.StopAtDistance(7)))

		// Try to calculate based on a square boundary around the seal which corner is safer, then tele there
		//actions = append(actions, action.NewStepChain(func(d data.Data) []step.Step {
		//	if obj, found := d.Objects.FindOne(seal); found {
		//		pos := a.getLessConcurredCornerAroundSeal(d, obj.Position)
		//		return []step.Step{step.MoveTo(pos)}
		//	}
		//	return []step.Step{}
		//}))

		// Kill all the monsters close to the seal and item pickup
		//actions = append(actions,
		//	a.builder.ClearAreaAroundPlayer(13),
		//	a.builder.ItemPickup(false, 40),
		//)

		// Activate the seal
		actions = append(actions,
			a.builder.ClearAreaAroundPlayer(15),
			action.NewStepChain(func(d data.Data) []step.Step {
				a.logger.Debug("Trying to activate seal...", zap.Int("seal", sealNumber+1))
				lastInteractionAt := time.Now()
				return []step.Step{
					step.InteractObject(seal, func(d data.Data) bool {
						if obj, found := d.Objects.FindOne(seal); found {
							if !obj.Selectable {
								a.logger.Debug("Seal activated, waiting for elite group to spawn", zap.Int("seal", sealNumber+1))
							}
							return !obj.Selectable
						}
						if time.Since(lastInteractionAt) > time.Second*3 {
							lastInteractionAt = time.Now()
							pather.RandomMovement()
							time.Sleep(time.Second)
						}
						return false
					}),
				}
			}),
		)

		// Only if we are not in the first seal
		if sealNumber != 0 {
			if sealNumber == 2 {
				// TODO: Refactor this thing...
				// Now wait & try to kill the Elite packs (maybe are already dead, killed during previous action)
				startTime := time.Time{}
				found := false
				actions = append(actions, action.NewStepChain(func(d data.Data) []step.Step {
					if startTime.IsZero() {
						startTime = time.Now()
					}
					for _, m := range d.Monsters.Enemies(data.MonsterEliteFilter()) {
						if a.isSealElite(m) {
							found = true
							a.logger.Debug("Seal defender found!")
							return nil
						}
					}

					if time.Since(startTime) < time.Second*4 {
						return []step.Step{step.Wait(time.Millisecond * 100)}
					}

					return nil
				}, action.RepeatUntilNoSteps()))

				if !found {
					actions = append(actions, a.builder.MoveToCoords(data.Position{
						X: 7785,
						Y: 5237,
					}))
				}
			} else {
				// First clear close trash mobs, regardless if they are elite or not
				actions = append(actions, a.builder.ClearAreaAroundPlayer(15))
			}

			// Now wait & try to kill the Elite packs (maybe are already dead, killed during previous action)
			startTime := time.Time{}
			actions = append(actions, action.NewStepChain(func(d data.Data) []step.Step {
				if startTime.IsZero() {
					startTime = time.Now()
				}
				for _, m := range d.Monsters.Enemies(data.MonsterEliteFilter()) {
					if a.isSealElite(m) {
						a.logger.Debug("Seal defender found!")
						return nil
					}
				}

				if time.Since(startTime) < time.Second*5 {
					return []step.Step{step.Wait(time.Millisecond * 100)}
				}

				return nil
			}, action.RepeatUntilNoSteps()))

			actions = append(actions, a.char.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
				for _, m := range d.Monsters.Enemies(data.MonsterEliteFilter()) {
					if a.isSealElite(m) {
						_, _, found := pather.GetPath(d, m.Position)
						return m.UnitID, found
					}
				}

				return 0, false
			}, nil))
		}

		actions = append(actions, a.builder.ItemPickup(false, 40))
	}

	// Go back to town to buy potions if needed
	actions = append(actions, action.NewChain(func(d data.Data) []action.Action {
		_, isLevelingChar := a.char.(action.LevelingCharacter)
		if isLevelingChar && (a.bm.ShouldBuyPotions(d) || (config.Config.Character.UseMerc && d.MercHPPercent() <= 0)) {
			return a.builder.InRunReturnTownRoutine()
		}

		return nil
	}))

	actions = append(actions,
		a.builder.Buff(),
		a.builder.MoveToCoords(diabloSpawnPosition),
		a.char.KillDiablo(),
	)

	return
}

func (a Diablo) getLessConcurredCornerAroundSeal(d data.Data, sealPosition data.Position) data.Position {
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
			fmt.Printf("Moving to corner %d. Has no monsters.\n", i)
			return corners[i]
		}
		fmt.Printf("Corner %d. Average monster distance: %d\n", i, averageDistance)
	}

	fmt.Printf("Moving to corner %d. Average monster distance: %d\n", bestCorner, bestCornerDistance)

	return corners[bestCorner]
}

func (a Diablo) isSealElite(monster data.Monster) bool {
	if monster.Type == data.MonsterTypeSuperUnique && (monster.Name == npc.OblivionKnight || monster.Name == npc.VenomLord || monster.Name == npc.StormCaster) {
		return true
	}

	return false
}
