package run

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type Pit struct {
	baseRun
}

func (a Pit) Name() string {
	return "Pit"
}

func (a Pit) BuildActions() (actions []action.Action) {
	// Moving to starting point (OuterCloister)
	actions = append(actions, a.builder.WayPoint(area.BlackMarsh))

	// Buff
	actions = append(actions, a.char.Buff())

	// move to TamoeHighland
	actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
		return []step.Step{
			// step.MoveToLevel(area.MonasteryGate),
			step.MoveToLevel(area.TamoeHighland),
		}
	}))

	// Travel to pit level 1
	actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
		fmt.Println(fmt.Sprintf("Travel to pit level 1"))
		return []step.Step{
			step.MoveToLevel(area.PitLevel1),
			step.SyncStep(func(_ data.Data) error {
				// Add small delay to fetch the monsters
				helper.Sleep(1000)
				return nil
			}),
		}
	}))

	if config.Config.Companion.Enabled && config.Config.Companion.Leader {
		actions = append(actions, action.BuildStatic(func(_ data.Data) []step.Step {
			return []step.Step{step.OpenPortal()}
		}))
	}

	// Clear pit level 1
	actions = append(actions, a.builder.ClearArea(true, data.MonsterAnyFilter()))

	// Travel to pit level 2
	actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(area.PitLevel2),
			step.SyncStep(func(_ data.Data) error {
				// Add small delay to fetch the monsters
				helper.Sleep(1000)
				return nil
			}),
		}
	}))

	// Clear pit level 2
	actions = append(actions, a.builder.ClearArea(true, data.MonsterAnyFilter()))

	return
}
