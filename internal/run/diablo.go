package run

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/npc"
	"github.com/hectorgimenez/koolo/internal/game/object"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"go.uber.org/zap"
)

const (
	chaosSanctuaryEntranceX = 7800
	chaosSanctuaryEntranceY = 5600

	diabloSpawnX = 7800
	diabloSpawnY = 5286
)

type Diablo struct {
	baseRun
}

func (a Diablo) Name() string {
	return "Diablo"
}

func (a Diablo) BuildActions() (actions []action.Action) {
	// Moving to starting point (RiverOfFlame)
	actions = append(actions, a.builder.WayPoint(area.RiverOfFlame))

	// Buff
	actions = append(actions, a.char.Buff())

	// Travel to diablo spawn location
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveTo(chaosSanctuaryEntranceX, chaosSanctuaryEntranceY, true),

			// This hacky thing is forcing the player to move to the new area, because Pather can not [yet]
			// walk through different areas without a game loading screen.
			step.SyncStepWithCheck(func(data game.Data) error {
				hid.KeyDown(config.Config.Bindings.ForceMove)
				helper.Sleep(500)
				hid.KeyUp(config.Config.Bindings.ForceMove)

				return nil
			}, func(data game.Data) step.Status {
				if data.PlayerUnit.Area == area.ChaosSanctuary {
					return step.StatusCompleted
				}

				return step.StatusInProgress
			}),

			// We move to the center in order to load all the seals in memory
			step.MoveTo(diabloSpawnX, diabloSpawnY, true),
		}
	}))

	seals := []object.Name{object.DiabloSeal4, object.DiabloSeal5, object.DiabloSeal3, object.DiabloSeal2, object.DiabloSeal1}

	// Move across all the seals, try to find the most clear spot around them, kill monsters and activate the seal.
	for i, s := range seals {
		seal := s
		sealNumber := i
		// Go to the seal and stop at distance 30, close enough to fetch monsters from memory
		actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
			a.logger.Debug("Moving to next seal", zap.Int("seal", sealNumber+1))
			if obj, found := data.Objects.FindOne(seal); found {
				a.logger.Debug("Seal found, start teleporting", zap.Int("seal", sealNumber+1))
				return []step.Step{step.MoveTo(obj.Position.X, obj.Position.Y, true, step.StopAtDistance(30))}
			}
			a.logger.Debug("SEAL NOT FOUND", zap.Int("seal", sealNumber+1))
			return []step.Step{}
		}))

		// Try to calculate based on a square boundary around the seal which corner is safer, then tele there
		actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
			if obj, found := data.Objects.FindOne(seal); found {
				pos := a.getLessConcurredCornerAroundSeal(data, obj.Position)
				return []step.Step{step.MoveTo(pos.X, pos.Y, true)}
			}
			return []step.Step{}
		}))

		// Kill all the monsters close to the seal and item pickup
		actions = append(actions, a.builder.ClearAreaAroundPlayer(13))
		actions = append(actions, a.builder.ItemPickup(false, 40))

		// Activate the seal
		actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
			return []step.Step{
				step.InteractObject(seal, func(data game.Data) bool {
					if obj, found := data.Objects.FindOne(seal); found {
						return !obj.Selectable
					}
					return false
				}),
			}
		}))

		// First clear close trash mobs, regardless if they are elite or not
		actions = append(actions, a.builder.ClearAreaAroundPlayer(10))

		// Now try to kill the Elite packs (maybe are already dead)
		actions = append(actions, a.char.KillMonsterSequence(func(data game.Data) (game.UnitID, bool) {
			for _, m := range data.Monsters.Enemies(game.MonsterEliteFilter()) {
				if a.isSealElite(m) {
					a.logger.Debug("FOUND SEAL DEFENDER!!!")
					return m.UnitID, true
				}
			}

			return 0, false
		}, nil))

		actions = append(actions, a.builder.ItemPickup(false, 40))
	}

	return
}

func (a Diablo) getLessConcurredCornerAroundSeal(data game.Data, sealPosition game.Position) game.Position {
	corners := [4]game.Position{
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
		for _, m := range data.Monsters.Enemies() {
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

func (a Diablo) isSealElite(monster game.Monster) bool {
	switch monster.Type {
	case game.MonsterTypeSuperUnique:
		switch monster.Name {
		case npc.OblivionKnight, npc.VenomLord, npc.StormCaster:
			return true
		}
	case game.MonsterTypeMinion:
		switch monster.Name {
		case npc.DoomKnight, npc.VenomLord, npc.StormCaster:
			return true
		}
	}

	return false
}
