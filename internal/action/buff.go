package action

import (
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"

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

		preKeys := make([]data.KeyBinding, 0)
		for _, buff := range b.ch.PreCTABuffSkills(d) {
			kb, found := d.KeyBindings.KeyBindingForSkill(buff)
			if !found {
				b.Logger.Info("Key binding not found, skipping buff", slog.String("skill", buff.Desc().Name))
			} else {
				preKeys = append(preKeys, kb)
			}
		}

		if len(preKeys) > 0 {
			b.Logger.Debug("PRE CTA Buffing...")
			steps = append(steps,
				step.SyncStep(func(_ game.Data) error {
					for _, kb := range preKeys {
						helper.Sleep(100)
						b.HID.PressKeyBinding(kb)
						helper.Sleep(180)
						b.HID.Click(game.RightButton, 640, 340)
						helper.Sleep(100)
					}
					return nil
				}),
			)
		}

		steps = append(steps, b.buffCTA(d)...)

		postKeys := make([]data.KeyBinding, 0)
		for _, buff := range b.ch.BuffSkills(d) {
			kb, found := d.KeyBindings.KeyBindingForSkill(buff)
			if !found {
				b.Logger.Info("Key binding not found, skipping buff", slog.String("skill", buff.Desc().Name))
			} else {
				postKeys = append(postKeys, kb)
			}
		}

		if len(postKeys) > 0 {
			b.Logger.Debug("Post CTA Buffing...")

			steps = append(steps,
				step.SyncStep(func(_ game.Data) error {
					for _, kb := range postKeys {
						helper.Sleep(100)
						b.HID.PressKeyBinding(kb)
						helper.Sleep(180)
						b.HID.Click(game.RightButton, 640, 340)
						helper.Sleep(100)
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

	if b.ctaFound(d) && (!d.PlayerUnit.States.HasState(state.Battleorders) || !d.PlayerUnit.States.HasState(state.Battlecommand)) {
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
			if buff == skill.CycloneArmor && !d.PlayerUnit.States.HasState(state.Cyclonearmor) {
				return true
			}
		}
	}

	return false
}

func (b *Builder) buffCTA(d game.Data) (steps []step.Step) {
	if b.ctaFound(d) {
		b.Logger.Debug("CTA found: swapping weapon and casting Battle Command / Battle Orders")

		// Swap weapon only in case we don't have the CTA, sometimes CTA is already equipped (for example chicken previous game during buff stage)
		if _, found := d.PlayerUnit.Skills[skill.BattleCommand]; !found {
			steps = append(steps, step.SwapToCTA())
		}

		steps = append(steps,
			step.SyncStep(func(d game.Data) error {
				b.HID.PressKeyBinding(d.KeyBindings.MustKBForSkill(skill.BattleCommand))
				helper.Sleep(180)
				b.HID.Click(game.RightButton, 300, 300)
				helper.Sleep(100)
				b.HID.PressKeyBinding(d.KeyBindings.MustKBForSkill(skill.BattleOrders))
				helper.Sleep(180)
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

func (b *Builder) ctaFound(d game.Data) bool {
	for _, itm := range d.Inventory.ByLocation(item.LocationEquipped) {
		_, boFound := itm.FindStat(stat.NonClassSkill, int(skill.BattleOrders))
		_, bcFound := itm.FindStat(stat.NonClassSkill, int(skill.BattleCommand))

		if boFound && bcFound {
			return true
		}
	}

	return false
}
