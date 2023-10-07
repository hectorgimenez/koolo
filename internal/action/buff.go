package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
)

func (b *Builder) EnsureBuff(forceReapply bool) *StepChainAction {
	return NewStepChain(func(d data.Data) (steps []step.Step) {
		rebuffRequired := false
		buffs := b.ch.BuffSkills()
		for _, b := range buffs {
		}

		steps = append(steps, b.buffCTA(d)...)

		return append(steps,
			step.SyncStep(func(d data.Data) error {
				if _, found := d.PlayerUnit.Skills[skill.HolyShield]; !found {
					return nil
				}

				if config.Config.Bindings.Paladin.HolyShield != "" {
					hid.PressKey(config.Config.Bindings.Paladin.HolyShield)
					helper.Sleep(100)
					hid.Click(hid.RightButton)
				}

				return nil
			}),
		)
	})
}

func (b *Builder) buffCTA(d data.Data) (steps []step.Step) {
	ctaFound := false
	for _, itm := range d.Items.ByLocation(item.LocationEquipped) {
		if itm.Stats[stat.NumSockets].Value == 5 && itm.Stats[stat.ReplenishLife].Value == 12 && itm.Stats[stat.NonClassSkill].Value > 0 && itm.Stats[stat.PreventMonsterHeal].Value > 0 {
			ctaFound = true
			b.logger.Debug("CTA found: swapping weapon and casting Battle Command / Battle Orders")
			break
		}
	}

	if ctaFound && config.Config.Bindings.CTABattleCommand != "" && config.Config.Bindings.CTABattleOrders != "" {
		steps = append(steps,
			step.SwapWeapon(),
			step.SyncStep(func(d data.Data) error {
				hid.PressKey(config.Config.Bindings.CTABattleCommand)
				helper.Sleep(100)
				hid.Click(hid.RightButton)
				helper.Sleep(300)
				hid.PressKey(config.Config.Bindings.CTABattleOrders)
				helper.Sleep(100)
				hid.Click(hid.RightButton)
				helper.Sleep(100)

				return nil
			}),
			step.SwapWeapon(),
		)
	}

	return
}
