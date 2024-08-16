package step

import (
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/v2/context"
)

func SetSkill(id skill.ID) {
	ctx := context.Get()
	if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(id); found {
		if ctx.Data.PlayerUnit.RightSkill != id {
			ctx.HID.PressKeyBinding(kb)
		}
	}
}
