package action

import (
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func BuffIfRequired() {
	ctx := context.Get()

	if !IsRebuffRequired() || ctx.Data.PlayerUnit.Area.IsTown() {
		return
	}

	// Don't buff if we have 2 or more monsters close to the character.
	// Don't merge with the previous if, because we want to avoid this expensive check if we don't need to buff
	closeMonsters := 0
	for _, m := range ctx.Data.Monsters {
		if ctx.PathFinder.DistanceFromMe(m.Position) < 15 {
			closeMonsters++
		}
		// cheaper to check here and end function if say first 2 already < 15
		// so no need to compute the rest
		if closeMonsters >= 2 {
			return
		}
	}

	Buff()
}

func Buff() {
	ctx := context.Get()
	ctx.SetLastAction("Buff")

	if ctx.Data.PlayerUnit.Area.IsTown() || time.Since(ctx.LastBuffAt) < time.Second*30 {
		return
	}

	// Check if we're in loading screen
	if ctx.GameReader.GetData().OpenMenus.LoadingScreen {
		ctx.Logger.Debug("Loading screen detected. Waiting for game to load before buffing...")
		ctx.WaitForGameToLoad()

		// Give it half a second more
		utils.Sleep(500)
	}

	preKeys := make([]data.KeyBinding, 0)
	for _, buff := range ctx.Char.PreCTABuffSkills() {
		kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(buff)
		if !found {
			ctx.Logger.Info("Key binding not found, skipping buff", slog.String("skill", buff.Desc().Name))
		} else {
			preKeys = append(preKeys, kb)
		}
	}

	if len(preKeys) > 0 {
		ctx.Logger.Debug("PRE CTA Buffing...")
		for _, kb := range preKeys {
			utils.Sleep(100)
			ctx.HID.PressKeyBinding(kb)
			utils.Sleep(180)
			ctx.HID.Click(game.RightButton, 640, 340)
			utils.Sleep(100)
		}
	}

	buffCTA()

	postKeys := make([]data.KeyBinding, 0)
	for _, buff := range ctx.Char.BuffSkills() {
		kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(buff)
		if !found {
			ctx.Logger.Info("Key binding not found, skipping buff", slog.String("skill", buff.Desc().Name))
		} else {
			postKeys = append(postKeys, kb)
		}
	}

	if len(postKeys) > 0 {
		ctx.Logger.Debug("Post CTA Buffing...")

		for _, kb := range postKeys {
			utils.Sleep(100)
			ctx.HID.PressKeyBinding(kb)
			utils.Sleep(180)
			ctx.HID.Click(game.RightButton, 640, 340)
			utils.Sleep(100)
		}
		ctx.LastBuffAt = time.Now()
	}
}

var buffStateMap = map[skill.ID][]state.State{
	// map of buff skills and the related states to watch for to indicate that we need a rebuff
	skill.HolyShield: {state.State(state.Holyshield)},
	skill.FrozenArmor: {
		state.State(state.Frozenarmor),
		state.State(state.Shiverarmor),
		state.State(state.Chillingarmor),
	},
	skill.EnergyShield: {state.State(state.Energyshield)},
	skill.CycloneArmor: {state.State(state.Cyclonearmor)},
	skill.Fade:         {state.State(state.Fade)},
	// skill.BurstOfSpeed: {state.State(state.BurstOfSpeed)}, Not mapped?

}

func needsRebuff(buff skill.ID, ctx context.Status) bool {
	hasState := false
	if _, found := ctx.Data.KeyBindings.KeyBindingForSkill(buff); found {
		stateMappings, ok := buffStateMap[buff]
		if ok {
			for _, neededState := range stateMappings {
				if ctx.Data.PlayerUnit.States.HasState(neededState) {
					hasState = true
				}
			}
		} else {
			ctx.Logger.Error("Tried to buff with unimplemented buff state", slog.Any("skillID", buff))
			return false
		}
	}
	return !hasState
}

func IsRebuffRequired() bool {
	ctx := context.Get()
	ctx.SetLastAction("IsRebuffRequired")

	// Don't buff if we are in town, or we did it recently (it prevents double buffing because of network lag)
	if ctx.Data.PlayerUnit.Area.IsTown() || time.Since(ctx.LastBuffAt) < time.Second*30 {
		return false
	}

	if ctaFound(*ctx.Data) && (!ctx.Data.PlayerUnit.States.HasState(state.Battleorders) || !ctx.Data.PlayerUnit.States.HasState(state.Battlecommand)) {
		return true
	}

	buffs := ctx.Char.BuffSkills()

	rebuffRequired := false
	for _, buff := range buffs {
		rebuffRequired = needsRebuff(buff, *ctx)
		if rebuffRequired {
			return true
		}
	}

	return false
}

func buffCTA() {
	ctx := context.Get()
	ctx.SetLastAction("buffCTA")

	if ctaFound(*ctx.Data) {
		ctx.Logger.Debug("CTA found: swapping weapon and casting Battle Command / Battle Orders")

		// Swap weapon only in case we don't have the CTA, sometimes CTA is already equipped (for example chicken previous game during buff stage)
		if _, found := ctx.Data.PlayerUnit.Skills[skill.BattleCommand]; !found {
			step.SwapToCTA()
		}

		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(skill.BattleCommand))
		utils.Sleep(180)
		ctx.HID.Click(game.RightButton, 300, 300)
		utils.Sleep(180)
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(skill.BattleOrders))
		utils.Sleep(180)
		ctx.HID.Click(game.RightButton, 300, 300)
		utils.Sleep(100)

		utils.Sleep(500)
		step.SwapToMainWeapon()
	}
}

func ctaFound(d game.Data) bool {
	for _, itm := range d.Inventory.ByLocation(item.LocationEquipped) {
		_, boFound := itm.FindStat(stat.NonClassSkill, int(skill.BattleOrders))
		_, bcFound := itm.FindStat(stat.NonClassSkill, int(skill.BattleCommand))

		if boFound && bcFound {
			return true
		}
	}

	return false
}
