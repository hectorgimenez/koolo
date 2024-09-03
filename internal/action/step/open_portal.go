package step

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func OpenPortal() error {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "OpenPortal"

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
