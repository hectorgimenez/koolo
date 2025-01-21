package run

import (
	"fmt"
	"slices"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/utils"
)

type Mephisto struct {
	ctx                *context.Status
	clearMonsterFilter data.MonsterFilter // Used to clear area (basically TZ)
}

func NewMephisto(tzClearFilter data.MonsterFilter) *Mephisto {
	return &Mephisto{
		ctx:                context.Get(),
		clearMonsterFilter: tzClearFilter,
	}
}

func (m Mephisto) Name() string {
	return string(config.MephistoRun)
}

func (m Mephisto) Run() error {

	// Use waypoint to DuranceOfHateLevel2
	err := action.WayPoint(area.DuranceOfHateLevel2)
	if err != nil {
		return err
	}

	if m.clearMonsterFilter != nil {
		if err = action.ClearCurrentLevel(m.ctx.CharacterCfg.Game.Mephisto.OpenChests, m.clearMonsterFilter); err != nil {
			return err
		}
	}

	// Move to DuranceOfHateLevel3
	if err = action.MoveToArea(area.DuranceOfHateLevel3); err != nil {
		return err
	}

	// Move to the Safe position
	action.MoveToCoords(data.Position{
		X: 17568,
		Y: 8069,
	})

	// Disable item pickup while fighting Mephisto (prevent picking up items if nearby monsters die)
	m.ctx.DisableItemPickup()

	// Kill Mephisto
	err = m.ctx.Char.KillMephisto()

	// Enable item pickup after the fight
	m.ctx.EnableItemPickup()

	if err != nil {
		return err
	}

	if m.ctx.CharacterCfg.Game.Mephisto.OpenChests && m.ctx.CharacterCfg.Game.Mephisto.KillCouncilMembers {
		// Clear the area with the selected options
		return action.ClearCurrentLevel(m.ctx.CharacterCfg.Game.Mephisto.OpenChests, m.CouncilMemberFilter())

	} else if m.ctx.CharacterCfg.Game.Mephisto.OpenChests {

		// determine chests wanted, need to add stand/rack option to config
		interactableChests := []object.Name{181, 183, 104, 105, 106, 107}

		// find chests and racks
		var crs []data.Object
		for _, o := range m.ctx.Data.Objects {
			if slices.Contains(interactableChests, o.Name) {
				m.ctx.Logger.Debug("Found chest/rack at:", "position", o.Position)
				crs = append(crs, o)
			}
		}
		m.ctx.Logger.Debug("Total chests/racks found", "count", len(crs))

		// Handle each of the chests/racks one by one
		for _, crIndividual := range crs {
			// Move to the chest/rack
			err = action.MoveToCoords(crIndividual.Position)
			if err != nil {
				return err
			}

			// Interact with the object
			err = action.InteractObject(crIndividual, func() bool {
				object, _ := m.ctx.Data.Objects.FindByID(crIndividual.ID)
				return !object.Selectable
			})
			if err != nil {
				m.ctx.Logger.Warn(fmt.Sprintf("[%s] failed interacting with object [%v] in Area: [%s]", m.ctx.Name, crIndividual.Name, m.ctx.Data.PlayerUnit.Area.Area().Name), err)
			}
			utils.Sleep(500) // Add small delay to allow the game to open the object and drop the content
		}
	}

	return nil
}

func (m Mephisto) CouncilMemberFilter() data.MonsterFilter {
	return func(m data.Monsters) []data.Monster {
		var filteredMonsters []data.Monster
		for _, mo := range m {
			if mo.Name == npc.CouncilMember || mo.Name == npc.CouncilMember2 || mo.Name == npc.CouncilMember3 {
				filteredMonsters = append(filteredMonsters, mo)
			}
		}

		return filteredMonsters
	}
}
