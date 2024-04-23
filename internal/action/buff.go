package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
)

var lastBuffedAt = map[string]time.Time{}

func (b *Builder) BuffIfRequired(d game.Data) *StepChainAction {
	if !b.IsRebuffRequired(d) {
		return nil
	}

	// Don't buff if we have 2 or more monsters close to the character.
	// Don't merge with the previous if, because we want to avoid this expensive check if we don't need to buff
	closeMonsters := 0
	for _, m := range d.Monsters {
		if pather.DistanceFromMe(d, m.Position) < 15 {
			closeMonsters++
		}
	}
	if closeMonsters >= 2 {
		return nil
	}

	return b.Buff()
}

func getLastBuffedAt(supervisor string) time.Time {
	if t, found := lastBuffedAt[supervisor]; found {
		return t
	}
	return time.Time{}
}

func (b *Builder) Buff() *StepChainAction {
	return NewStepChain(func(d game.Data) (steps []step.Step) {
		if d.PlayerUnit.Area.IsTown() || time.Since(getLastBuffedAt(b.Supervisor)) < time.Second*30 {
			return nil
		}

		steps = append(steps, b.buffCTA(d)...)

		keys := make([]data.KeyBinding, 0)
		for _, buff := range b.ch.BuffSkills(d) {
			kb, found := d.KeyBindings.KeyBindingForSkill(buff)
			if !found {
				return nil
			}

			keys = append(keys, kb)
		}

		if len(keys) > 0 {
			b.Logger.Debug("Buffing...")

			steps = append(steps,
				step.SyncStep(func(_ game.Data) error {
					for _, kb := range keys {
						helper.Sleep(200)
						b.HID.PressKeyBinding(kb)
						helper.Sleep(300)
						b.HID.Click(game.RightButton, 300, 300)
						helper.Sleep(300)
					}
					return nil
				}),
			)
			lastBuffedAt[b.Supervisor] = time.Now()
		}

		return steps
	})
}

func (b *Builder) IsRebuffRequired(d game.Data) bool {
	// Don't buff if we are in town, or we did it recently (it prevents double buffing because of network lag)
	if d.PlayerUnit.Area.IsTown() || time.Since(getLastBuffedAt(b.Supervisor)) < time.Second*30 {
		return false
	}

	if b.isCTAEnabled(d) && (!d.PlayerUnit.States.HasState(state.Battleorders) || !d.PlayerUnit.States.HasState(state.Battlecommand)) {
		return true
	}

	// TODO: Find a better way to convert skill to state
	buffs := b.ch.BuffSkills(d)
	for _, buff := range buffs {
		if _, found := d.KeyBindings.KeyBindingForSkill(buff); found {
			if buff == skill.HolyShield && !d.PlayerUnit.States.HasState(state.Holyshield) {
				return true
			}
			if buff == skill.FrozenArmor && (!d.PlayerUnit.States.HasState(state.Frozenarmor) && !d.PlayerUnit.States.HasState(state.Shiverarmor) && !d.PlayerUnit.States.HasState(state.Chillingarmor)) {
				return true
			}
			if buff == skill.EnergyShield && !d.PlayerUnit.States.HasState(state.Energyshield) {
				return true
			}
		}
	}

	return false
}

func (b *Builder) buffCTA(d game.Data) (steps []step.Step) {
	if b.isCTAEnabled(d) {
		b.Logger.Debug("CTA found: swapping weapon and casting Battle Command / Battle Orders")

		// Swap weapon only in case we don't have the CTA, sometimes CTA is already equipped (for example chicken previous game during buff stage)
		if _, found := d.PlayerUnit.Skills[skill.BattleCommand]; !found {
			steps = append(steps, step.SwapToCTA())
		}

		steps = append(steps,
			step.SyncStep(func(d game.Data) error {
				b.HID.PressKeyBinding(d.KeyBindings.MustKBForSkill(skill.BattleCommand))
				helper.Sleep(100)
				b.HID.Click(game.RightButton, 300, 300)
				helper.Sleep(300)
				b.HID.PressKeyBinding(d.KeyBindings.MustKBForSkill(skill.BattleOrders))
				helper.Sleep(100)
				b.HID.Click(game.RightButton, 300, 300)
				helper.Sleep(100)

				return nil
			}),
			step.Wait(time.Millisecond*500),
			step.SwapToMainWeapon(),
		)
	}

	return
}

func (b *Builder) isCTAEnabled(d game.Data) bool {
	for _, itm := range d.Items.ByLocation(item.LocationEquipped) {
		if itm.Stats[stat.NumSockets].Value == 5 && itm.Stats[stat.ReplenishLife].Value == 12 && itm.Stats[stat.NonClassSkill].Value > 0 && itm.Stats[stat.PreventMonsterHeal].Value > 0 {
			return true
		}
	}

	return false
}
