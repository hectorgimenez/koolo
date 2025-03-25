package action

import (
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
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

func castSkill(buff skill.ID) {
	ctx := context.Get()
	kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(buff)

	if !found {
		ctx.Logger.Info("Key binding not found, skipping buff", slog.String("skill", buff.Desc().Name))
		return
	}
	utils.Sleep(100)
	ctx.HID.PressKeyBinding(kb)
	utils.Sleep(180)
	ctx.HID.Click(game.RightButton, 640, 340)
	utils.Sleep(100)
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

	ctx.Logger.Debug("Pre CTA Buffing...")
	for _, buff := range ctx.Char.PreCTABuffSkills() {
		castSkill(buff)
	}

	hasCTA := ctaFound(*ctx.Data)
	if hasCTA {
		if ctx.CharacterCfg.Character.BuffWithCTA {
			buffCTA(ctx.Char.BuffSkills())
		} else {
			buffCTA([]skill.ID{})
		}
	}

	ctx.Logger.Debug("Post CTA Buffing...")
	if !hasCTA || !ctx.CharacterCfg.Character.BuffWithCTA {
		for _, buff := range ctx.Char.BuffSkills() {
			castSkill(buff)
		}
	}

	ctx.LastBuffAt = time.Now()
}

var buffStateMap = map[skill.ID]state.State{
	// map of buff skills and the related state to watch for to indicate that we need a rebuff
	skill.HolyShield:    state.State(state.Holyshield),
	skill.FrozenArmor:   state.State(state.Frozenarmor),
	skill.ShiverArmor:   state.State(state.Shiverarmor),
	skill.ChillingArmor: state.State(state.Chillingarmor),
	skill.EnergyShield:  state.State(state.Energyshield),
	skill.CycloneArmor:  state.State(state.Cyclonearmor),
	skill.Fade:          state.State(state.Fade),
	skill.BurstOfSpeed:  state.State(state.Quickness),
	skill.BattleOrders:  state.State(state.Battleorders),
	skill.BattleCommand: state.State(state.Battlecommand),
}

func skillNeedsRebuff(buff skill.ID) bool {
	ctx := context.Get()
	hasState := false
	if _, found := ctx.Data.KeyBindings.KeyBindingForSkill(buff); found {
		neededState, ok := buffStateMap[buff]
		if ok {
			if ctx.Data.PlayerUnit.States.HasState(neededState) {
				hasState = true
			}
		} else {
			ctx.Logger.Error("Tried to buff with unimplemented buff state", slog.Any("Buff", buff.Desc().Name))
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

	if ctaFound(*ctx.Data) && (skillNeedsRebuff(skill.BattleOrders) || skillNeedsRebuff(skill.BattleCommand)) {
		return true
	}

	buffs := ctx.Char.BuffSkills()

	rebuffRequired := false
	for _, buff := range buffs {
		rebuffRequired = skillNeedsRebuff(buff)
		if rebuffRequired {
			return true
		}
	}

	return false
}

func buffCTA(otherBuffs []skill.ID) {
	ctx := context.Get()
	ctx.SetLastAction("buffCTA")

	ctx.Logger.Debug("CTA found: swapping weapon and casting Battle Command / Battle Orders")

	step.SwapToSecondary()

	yells := []skill.ID{
		skill.BattleCommand,
		skill.BattleOrders,
	}

	for _, yell := range yells {
		castSkill(yell)
	}

	// If applicable, cast other buffs while we're holding CTA
	for _, buff := range otherBuffs {
		castSkill(buff)
	}

	utils.Sleep(500)
	step.SwapToMainWeapon()
}

func ctaFound(d game.Data) bool {
	for _, itm := range d.Inventory.ByLocation(item.LocationEquipped) {
		ctaFound := itm.RunewordName == item.RunewordCallToArms
		if ctaFound {
			return true
		}
	}

	return false
}
