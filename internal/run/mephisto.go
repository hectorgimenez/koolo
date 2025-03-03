package run

import (
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
	action.OpenTPIfLeader()
	if m.clearMonsterFilter != nil {
		if err = action.ClearCurrentLevel(m.ctx.CharacterCfg.Game.Mephisto.OpenChests, m.clearMonsterFilter); err != nil {
			return err
		}
	}

	// Move to DuranceOfHateLevel3
	if err = action.MoveToArea(area.DuranceOfHateLevel3); err != nil {
		return err
	}
	action.OpenTPIfLeader()
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

	if m.ctx.CharacterCfg.Game.Mephisto.OpenChests || m.ctx.CharacterCfg.Game.Mephisto.KillCouncilMembers {
		// Clear the area with the selected options
		return action.ClearCurrentLevel(m.ctx.CharacterCfg.Game.Mephisto.OpenChests, m.CouncilMemberFilter())
	}

	if m.ctx.CharacterCfg.Game.Mephisto.ExitToA4 {
		m.ctx.Logger.Debug("Moving to bridge")
		action.MoveToCoords(data.Position{X: 17588, Y: 8068})
		//Wait for bridge to rise
		utils.Sleep(1000)

		m.ctx.Logger.Debug("Moving to red portal")
		portal, _ := m.ctx.Data.Objects.FindOne(object.HellGate)
		action.MoveToCoords(portal.Position)

		action.InteractObject(portal, func() bool {
			return m.ctx.Data.PlayerUnit.Area == area.ThePandemoniumFortress
		})
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
