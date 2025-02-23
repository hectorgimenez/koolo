package step

import (
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func OpenPortal() error {
	ctx := context.Get()
	ctx.SetLastStep("OpenPortal")

	lastRun := time.Time{}
	for {
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

		_, found := ctx.Data.Objects.FindOne(object.TownPortal)
		if found {
			return nil
		}

		// Give some time to portal to popup before retrying...
		if time.Since(lastRun) < time.Second*2 {
			continue
		}

		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(skill.TomeOfTownPortal))
		utils.Sleep(250)
		ctx.HID.Click(game.RightButton, 300, 300)
		lastRun = time.Now()
	}
}

func OpenNewPortal() error {
	ctx := context.Get()
	ctx.SetLastStep("OpenPortal")
	ctx.Logger.Info("Checking for a self owned portal nearby.")
	lastRun := time.Time{}
	for {
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

		for _, o := range ctx.Data.Objects {
			if o.IsPortal() && o.Owner == ctx.Data.PlayerUnit.Name && pather.DistanceFromPoint(ctx.Data.PlayerUnit.Position, o.Position) < 15 {
				ctx.Logger.Info("self owned nearby portal found")
				return nil
			}
		}
		
		// Give some time to portal to popup before retrying...
		if time.Since(lastRun) < time.Second*2 {
			continue
		}
		ctx.Logger.Info("opening fresh portal")
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(skill.TomeOfTownPortal))
		utils.Sleep(250)
		ctx.HID.Click(game.RightButton, 300, 300)
		lastRun = time.Now()
	}
}
