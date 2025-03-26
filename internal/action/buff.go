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
	ctx.Logger.Debug("WEAPON CACHE CHECKPOINT",
		slog.Bool("cache_valid", ctx.WeaponBonusCache.IsValid),
		slog.Any("cache_data", ctx.WeaponBonusCache))

	if ctx.Data.PlayerUnit.Area.IsTown() || time.Since(ctx.LastBuffAt) < time.Second*30 {
		return
	}

	if ctx.Data.OpenMenus.LoadingScreen {
		ctx.Logger.Debug("Loading screen detected. Waiting for game to load before buffing...")
		ctx.WaitForGameToLoad()
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
		if ctx.WeaponBonusCache.IsValid {
			ctx.Logger.Debug("PRE CTA Buffing with Best Weapon Slot...")
			bestSlot := getBestWeaponSlot(ctx)
			if bestSlot != ctx.Data.ActiveWeaponSlot {
				ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
				utils.Sleep(200)
			}
		}
		for _, kb := range preKeys {
			utils.Sleep(200)
			ctx.HID.PressKeyBinding(kb)
			utils.Sleep(280)
			ctx.HID.Click(game.RightButton, 640, 340)
			utils.Sleep(200)
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

		if ctx.WeaponBonusCache.IsValid {
			ctx.Logger.Debug("POST CTA Buffing with Best Weapon Slot...")
			bestSlot := getBestWeaponSlot(ctx)
			if bestSlot != ctx.Data.ActiveWeaponSlot {
				ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
				utils.Sleep(200)
			}
		}

		for _, kb := range postKeys {
			utils.Sleep(200)
			ctx.HID.PressKeyBinding(kb)
			utils.Sleep(280)
			ctx.HID.Click(game.RightButton, 640, 340)
			utils.Sleep(200)
		}
		step.SwapToMainWeapon()

		ctx.LastBuffAt = time.Now()
	}
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

	// TODO: Find a better way to convert skill to state
	buffs := ctx.Char.BuffSkills()
	for _, buff := range buffs {
		if _, found := ctx.Data.KeyBindings.KeyBindingForSkill(buff); found {
			if buff == skill.HolyShield && !ctx.Data.PlayerUnit.States.HasState(state.Holyshield) {
				return true
			}
			if buff == skill.FrozenArmor && (!ctx.Data.PlayerUnit.States.HasState(state.Frozenarmor) && !ctx.Data.PlayerUnit.States.HasState(state.Shiverarmor) && !ctx.Data.PlayerUnit.States.HasState(state.Chillingarmor)) {
				return true
			}
			if buff == skill.EnergyShield && !ctx.Data.PlayerUnit.States.HasState(state.Energyshield) {
				return true
			}
			if buff == skill.CycloneArmor && !ctx.Data.PlayerUnit.States.HasState(state.Cyclonearmor) {
				return true
			}
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
		utils.Sleep(280)
		ctx.HID.Click(game.RightButton, 300, 300)
		utils.Sleep(200)
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(skill.BattleCommand))
		utils.Sleep(280)
		ctx.HID.Click(game.RightButton, 300, 300)
		utils.Sleep(200)
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(skill.BattleOrders))
		utils.Sleep(280)
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

func buildGearCache() {
	ctx := context.Get()
	ctx.WeaponBonusCache.IsValid = false

	currentSlot := ctx.Data.ActiveWeaponSlot

	if currentSlot != 0 {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
		ctx.RefreshGameData()
		utils.Sleep(500)
	}

	ctx.WeaponBonusCache.Slot1AllClassBonus = calculateAllPlusClassBonus(ctx)
	utils.Sleep(200)

	ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
	ctx.RefreshGameData()
	utils.Sleep(500)
	ctx.WeaponBonusCache.Slot2AllClassBonus = calculateAllPlusClassBonus(ctx)
	utils.Sleep(200)

	if ctx.Data.ActiveWeaponSlot != 0 {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
		utils.Sleep(200)
	}

	ctx.WeaponBonusCache.IsValid = true

	ctx.Logger.Debug("Weapon bonus cache built",
		slog.Int("slot1_bonus", ctx.WeaponBonusCache.Slot1AllClassBonus),
		slog.Int("slot2_bonus", ctx.WeaponBonusCache.Slot2AllClassBonus),
		slog.Bool("is_cache_valid", ctx.WeaponBonusCache.IsValid),
	)
}

// Here we calculate all the bonuses from the player's gear and skills
// and return the total bonus for the All +Skills (ID 127) and +AddClassSkills (ID 83) stats.
// Right now we ignore +SpecificSkills from Stat ID 97, until we build a UseCase for it.
func calculateAllPlusClassBonus(ctx *context.Status) int {
	total := 0

	allSkills := 0
	for layer := 0; layer < 10; layer++ {
		if s, found := ctx.Data.PlayerUnit.Stats.FindStat(stat.AllSkills, layer); found {
			allSkills += s.Value
		}
	}
	total += allSkills

	classSkills := 0
	classLayer := int(ctx.Data.PlayerUnit.Class)
	if s, found := ctx.Data.PlayerUnit.Stats.FindStat(stat.AddClassSkills, classLayer); found {
		classSkills = s.Value
	}
	total += classSkills

	ctx.Logger.Debug("Skill Bonus Calculation",
		slog.Int("all_skills", allSkills),
		slog.Int("class_skills", classSkills),
		slog.Int("total_bonus", total),
		slog.Int("class_id", int(ctx.Data.PlayerUnit.Class)),
	)

	return total
}

func getBestWeaponSlot(ctx *context.Status) int {
	if !ctx.WeaponBonusCache.IsValid {
		ctx.Logger.Debug("Cache invalid - using default slot 1")
		return 0
	}

	if ctx.WeaponBonusCache.Slot1AllClassBonus >= ctx.WeaponBonusCache.Slot2AllClassBonus {
		ctx.Logger.Debug("Selected slot 1 as best weapon slot",
			slog.Int("slot1_bonus", ctx.WeaponBonusCache.Slot1AllClassBonus),
			slog.Int("slot2_bonus", ctx.WeaponBonusCache.Slot2AllClassBonus),
		)
		return 0
	} else {
		ctx.Logger.Debug("Selected slot 2 as best weapon slot",
			slog.Int("slot1_bonus", ctx.WeaponBonusCache.Slot1AllClassBonus),
			slog.Int("slot2_bonus", ctx.WeaponBonusCache.Slot2AllClassBonus),
		)
		return 1
	}
}
