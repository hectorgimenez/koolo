package action

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) Repair() *StaticAction {
	return BuildStatic(func(data game.Data) (steps []step.Step) {
		durabilityPct := float32(data.PlayerUnit.Stats[stat.Durability]) / float32(data.PlayerUnit.Stats[stat.MaxDurability])
		if durabilityPct < 0.80 {
			b.logger.Info(fmt.Sprintf("Repairing, current durability: %0.2f is under 0.80", durabilityPct))

			x, y := int(float32(hid.GameAreaSizeX)/3.52), int(float32(hid.GameAreaSizeY)/1.37)

			steps = append(steps,
				step.InteractNPC(town.GetTownByArea(data.PlayerUnit.Area).RepairNPC()),
				step.KeySequence("home", "down", "enter"),
				step.SyncStep(func(_ game.Data) error {
					helper.Sleep(100)
					hid.MovePointer(x, y)
					hid.Click(hid.LeftButton)
					helper.Sleep(500)
					hid.PressKey("esc")
					return nil
				}),
			)
		}

		return
	}, Resettable(), CanBeSkipped())
}
