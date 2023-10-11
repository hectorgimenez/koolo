package action

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
)

func (b *Builder) Buff() *StepChainAction {
	return NewStepChain(func(d data.Data) (steps []step.Step) {
		if !b.IsRebuffRequired(d) {
			return nil
		}

		steps = append(steps, b.buffCTA(d)...)

		for buff, kb := range b.ch.BuffSkills() {
			if _, found := d.PlayerUnit.Skills[buff]; !found {
				return nil
			}

			if kb != "" {
				steps = append(steps,
					step.SyncStep(func(_ data.Data) error {
						hid.PressKey(kb)
						helper.Sleep(100)
						hid.Click(hid.RightButton)

						return nil
					}),
				)
			}
		}

		if len(steps) > 0 {
			b.logger.Debug("Buffing...")
			steps = append(steps, step.Wait(time.Millisecond*200)) // Small delay to let the game detect the buff
		}

		return steps
	})
}

func (b *Builder) IsRebuffRequired(d data.Data) bool {
	// Don't buff if we are in town
	if d.PlayerUnit.Area.IsTown() {
		return false
	}

	if b.isCTAEnabled(d) && (!d.PlayerUnit.States.HasState(state.Battleorders) || !d.PlayerUnit.States.HasState(state.Battlecommand)) {
		return true
	}

	// TODO: Find a better way to convert skill to state
	buffs := b.ch.BuffSkills()
	for buff, kb := range buffs {
		if kb != "" {
			if buff == skill.HolyShield && !d.PlayerUnit.States.HasState(state.Holyshield) {
				return true
			}
			if buff == skill.FrozenArmor && (!d.PlayerUnit.States.HasState(state.Frozenarmor) && !d.PlayerUnit.States.HasState(state.Shiverarmor) && !d.PlayerUnit.States.HasState(state.Chillingarmor)) {
				return true
			}
		}
	}

	return false
}

func (b *Builder) buffCTA(d data.Data) (steps []step.Step) {
	if b.isCTAEnabled(d) {
		b.logger.Debug("CTA found: swapping weapon and casting Battle Command / Battle Orders")

		// Swap weapon only in case we don't have the CTA, sometimes CTA is already equipped (for example chicken previous game during buff stage)
		if _, found := d.PlayerUnit.Skills[skill.BattleCommand]; !found {
			steps = append(steps, step.SwapWeapon())
		}

		steps = append(steps,
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

func (b *Builder) isCTAEnabled(d data.Data) bool {
	if config.Config.Bindings.CTABattleCommand == "" || config.Config.Bindings.CTABattleOrders == "" {
		return false
	}

	for _, itm := range d.Items.ByLocation(item.LocationEquipped) {
		if itm.Stats[stat.NumSockets].Value == 5 && itm.Stats[stat.ReplenishLife].Value == 12 && itm.Stats[stat.NonClassSkill].Value > 0 && itm.Stats[stat.PreventMonsterHeal].Value > 0 {
			return true
		}
	}

	return false
}
