package step

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
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

		tpTome, found := ctx.Data.Inventory.Find(item.TomeOfTownPortal, item.LocationInventory)
		if !found {
			return fmt.Errorf("no TP tome in inventory")
		}

		if st, statFound := tpTome.FindStat(stat.Quantity, 0); !statFound || st.Value == 0 {
			return fmt.Errorf("no TP charges in tome")
		}

		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(skill.TomeOfTownPortal))
		utils.Sleep(250)
		ctx.HID.Click(game.RightButton, 300, 300)
		lastRun = time.Now()
	}
}
