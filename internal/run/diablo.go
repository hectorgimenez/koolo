package run

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

var diabloSpawnPosition = data.Position{X: 7792, Y: 5294}
var chaosNavToPosition = data.Position{X: 7732, Y: 5292} //into path towards vizier
var sealPopChaosEntrace = data.Position{X: 7794, Y: 5558}

type Diablo struct {
	ctx *context.Status
}

func NewDiablo() *Diablo {
	return &Diablo{
		ctx: context.Get(),
	}
}

func (d *Diablo) Name() string {
	return string(config.DiabloRun)
}

func (d *Diablo) Run() error {
	// Just to be sure we always re-enable item pickup after the run
	defer func() {
		d.ctx.EnableItemPickup()
	}()

	if err := action.WayPoint(area.RiverOfFlame); err != nil {
		return err
	}

	action.MoveToArea(area.ChaosSanctuary)

	// We move directly to Diablo spawn position if StartFromStar is enabled, not clearing the path
	if d.ctx.CharacterCfg.Game.Diablo.StartFromStar {
		//move to star
		if err := action.MoveToCoords(diabloSpawnPosition); err != nil {
			return err
		}
		//open portal if leader
		if d.ctx.CharacterCfg.Companion.Leader {
			action.OpenTPIfLeader()
			action.Buff()
			action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())
		}
	} else {
		//open portal in entrance
		if d.ctx.CharacterCfg.Companion.Leader {
			action.OpenTPIfLeader()
			action.Buff()
			action.ClearAreaAroundPlayer(30, data.MonsterAnyFilter())
		}
		//path through towards vizier
		err := action.ClearThroughPath(chaosNavToPosition, 30, d.getMonsterFilter())
		if err != nil {
			return err
		}
	}

	sealGroups := map[string][]object.Name{
		"Vizier":       {object.DiabloSeal4, object.DiabloSeal5}, // Vizier
		"Lord De Seis": {object.DiabloSeal3},                     // Lord De Seis
		"Infector":     {object.DiabloSeal1, object.DiabloSeal2}, // Infector
	}

	// Thanks Go for the lack of ordered maps
	for _, bossName := range []string{"Vizier", "Lord De Seis", "Infector"} {
		d.ctx.Logger.Debug("Heading to", bossName)

		for _, sealID := range sealGroups[bossName] {
			seal, found := d.ctx.Data.Objects.FindOne(sealID)
			if !found {
				return fmt.Errorf("seal not found: %d", sealID)
			}

			err := action.ClearThroughPath(seal.Position, 20, d.getMonsterFilter())
			if err != nil {
				return err
			}

			// Handle the special case for DiabloSeal3
			if sealID == object.DiabloSeal3 && seal.Position.X == 7773 && seal.Position.Y == 5155 {
				if err = action.MoveToCoords(data.Position{X: 7768, Y: 5160}); err != nil {
					return fmt.Errorf("failed to move to bugged seal position: %w", err)
				}
			}

			// Check for seal pop config and if this is the last seal
			if d.ctx.CharacterCfg.Game.Diablo.SealPop && sealID == object.DiabloSeal2 {

				// Disable item pickup before traversing since we will be going back after seal pop
				d.ctx.DisableItemPickup()

				// Traverse the sanctuary to reactivate monsters
				if err = d.traverseSanctuary(d.ctx.Data.PlayerUnit.Position); err != nil {
					// Enable item pickup if traverse fails
					d.ctx.EnableItemPickup()
					return err
				}

				// Re enable item pickup on success
				d.ctx.EnableItemPickup()

			}

			// Clear everything around the seal
			action.ClearAreaAroundPlayer(10, d.ctx.Data.MonsterFilterAnyReachable())

			//Buff refresh before Infector
			if object.DiabloSeal1 == sealID {
				action.Buff()
			}

			if err = action.InteractObject(seal, func() bool {
				seal, _ = d.ctx.Data.Objects.FindOne(sealID)
				return !seal.Selectable
			}); err != nil {
				// Rerunning seal attempt with a random movement first before erroring out, seems to resolve most failed to interact
				d.ctx.PathFinder.RandomMovement()
				if err = action.InteractObject(seal, func() bool {
					seal, _ = d.ctx.Data.Objects.FindOne(sealID)
					return !seal.Selectable
				}); err != nil {
					return fmt.Errorf("failed to interact with seal: %w", err)
				}
			}

			// Traverse sanctuary to check for items that may have dropped from activated monsters
			if d.ctx.CharacterCfg.Game.Diablo.SealPop && sealID == object.DiabloSeal2 {

				// Make sure item pickup is enabled
				d.ctx.EnableItemPickup()

				// Traverse the sanctuary to loot reactivate monsters, end at Diablo
				if err = d.traverseSanctuary(diabloSpawnPosition); err != nil {
					return err
				}
			}

			// Infector spawns when first seal is enabled
			if object.DiabloSeal1 == sealID {
				if err = d.killSealElite(bossName); err != nil {
					return err
				}
			}
		}

		// Skip Infector boss because was already killed
		if bossName != "Infector" {
			// Wait for the boss to spawn and kill it.
			// Lord De Seis sometimes it's far, and we can not detect him, but we will kill him anyway heading to the next seal
			if err := d.killSealElite(bossName); err != nil && bossName != "Lord De Seis" {
				return err
			}
		}

	}

	if d.ctx.CharacterCfg.Game.Diablo.KillDiablo {
		// Bot will buff while seal popping, no need to force it if popping
		if !d.ctx.CharacterCfg.Game.Diablo.SealPop {
			action.Buff()
		}

		action.MoveToCoords(diabloSpawnPosition)

		// Check if we should disable item pickup for Diablo
		if d.ctx.CharacterCfg.Game.Diablo.DisableItemPickupDuringBosses {
			d.ctx.DisableItemPickup()
		}

		return d.ctx.Char.KillDiablo()
	}

	return nil
}

func (d *Diablo) killSealElite(boss string) error {
	d.ctx.Logger.Debug(fmt.Sprintf("Starting kill sequence for %s", boss))
	startTime := time.Now()
	timeout := 4 * time.Second

	for time.Since(startTime) < timeout {
		for _, m := range d.ctx.Data.Monsters.Enemies(d.ctx.Data.MonsterFilterAnyReachable()) {
			if action.IsMonsterSealElite(m) {
				d.ctx.Logger.Debug(fmt.Sprintf("Seal elite found: %s at position X: %d, Y: %d", m.Name, m.Position.X, m.Position.Y))

				return action.ClearAreaAroundPosition(m.Position, 30, d.ctx.Data.MonsterFilterAnyReachable())
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("no seal elite found for %s within %v seconds", boss, timeout.Seconds())
}

func (d *Diablo) getMonsterFilter() data.MonsterFilter {
	return func(monsters data.Monsters) (filteredMonsters []data.Monster) {
		for _, m := range monsters {
			if !d.ctx.Data.AreaData.IsWalkable(m.Position) {
				continue
			}

			// If FocusOnElitePacks is enabled, only return elite monsters and seal bosses
			if d.ctx.CharacterCfg.Game.Diablo.FocusOnElitePacks {
				if m.IsElite() || action.IsMonsterSealElite(m) {
					filteredMonsters = append(filteredMonsters, m)
				}
			} else {
				filteredMonsters = append(filteredMonsters, m)
			}
		}

		return filteredMonsters
	}
}

func (d *Diablo) traverseSanctuary(lastCoords data.Position) error {
	d.ctx.Logger.Debug("Traversing Chaos Sanctuary for Seal POP")

	// Move to sanctuary entrance
	if err := action.MoveToCoords(sealPopChaosEntrace); err != nil {
		return err
	}
	//Move back to last coords
	if err := action.MoveToCoords(lastCoords); err != nil {
		return err
	}

	return nil
}
