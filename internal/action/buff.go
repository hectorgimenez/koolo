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

	preCTASkills := ctx.Char.PreCTABuffSkills()
	if len(preCTASkills) > 0 {
		ctx.Logger.Debug("PRE CTA Buffing...")

		if ctx.WeaponBonusCache.IsValid {

			slot0Pre, slot1Pre := groupSkillsBySlot(ctx, preCTASkills)

			if len(slot0Pre) > 0 && ctx.Data.ActiveWeaponSlot != 0 {
				ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
				utils.Sleep(200)
				ctx.Logger.Debug("PRE CTA Buffing Slot 0...")
			}
			castSkills(ctx, slot0Pre)

			if len(slot1Pre) > 0 && ctx.Data.ActiveWeaponSlot != 1 {
				ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
				utils.Sleep(200)
				ctx.Logger.Debug("PRE CTA Buffing Slot 1...")
			}
			castSkills(ctx, slot1Pre)
		} else {

			castSkills(ctx, preCTASkills)
		}
	}

	buffCTA()

	postCTASkills := ctx.Char.BuffSkills()
	if len(postCTASkills) > 0 {
		ctx.Logger.Debug("Post CTA Buffing...")

		if ctx.WeaponBonusCache.IsValid {

			slot0Post, slot1Post := groupSkillsBySlot(ctx, postCTASkills)

			if len(slot0Post) > 0 && ctx.Data.ActiveWeaponSlot != 0 {
				ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
				utils.Sleep(200)
				ctx.Logger.Debug("Post CTA Buffing Slot 0...")
			}
			castSkills(ctx, slot0Post)

			if len(slot1Post) > 0 && ctx.Data.ActiveWeaponSlot != 1 {
				ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
				utils.Sleep(200)
				ctx.Logger.Debug("Post CTA Buffing Slot 1...")
			}
			castSkills(ctx, slot1Post)
		} else {
			castSkills(ctx, postCTASkills)
		}

		step.SwapToMainWeapon()

	}

	ctx.LastBuffAt = time.Now()
}

func groupSkillsBySlot(ctx *context.Status, skills []skill.ID) (slot0, slot1 []skill.ID) {
	for _, skillID := range skills {
		if ctx.WeaponBonusCache.OptimalSkillSlots[skillID] == 0 {
			slot0 = append(slot0, skillID)
		} else {
			slot1 = append(slot1, skillID)
		}
	}
	return
}

func castSkills(ctx *context.Status, skills []skill.ID) {
	cachedKeys := make(map[skill.ID]data.KeyBinding)

	for _, skillID := range skills {
		kb, ok := cachedKeys[skillID]
		if !ok {
			var found bool
			kb, found = ctx.Data.KeyBindings.KeyBindingForSkill(skillID)
			if !found {
				ctx.Logger.Info("Key binding not found", slog.String("skill", skillID.Desc().Name))
				continue
			}
			cachedKeys[skillID] = kb
		}

		utils.Sleep(200)
		ctx.HID.PressKeyBinding(kb)
		utils.Sleep(280)
		ctx.HID.Click(game.RightButton, 640, 340)
		utils.Sleep(200)
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

// buildGearCache, getspecificSkillBonuses, calculateSlotBonus, should now account for +allSkills ID 127 +classSkills ID 83 and +specificSkills ID 97
// I have not found a solution for adding +addskilltab ID 188 because skill.go does not have the according trees mapped
// I think currently ID 188 is giving me a value and a layer, and this layer corresponds to the skill tree, but I am not sure
// But even if i hardcode these layers to the skilltrees, it will not work because the skill.go does not have the trees mapped to the skills
func buildGearCache() {
	ctx := context.Get()
	ctx.WeaponBonusCache.IsValid = false

	currentSlot := ctx.Data.ActiveWeaponSlot
	if currentSlot != 0 {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
		ctx.RefreshGameData()
		utils.Sleep(500)
	}

	skills := append(ctx.Char.PreCTABuffSkills(), ctx.Char.BuffSkills()...)
	specificBonuses := getSpecificSkillBonuses(ctx, skills)

	ctx.WeaponBonusCache.Slot1AllClassBonus = calculateSlotBonus(ctx, 0, specificBonuses)
	utils.Sleep(200)

	ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
	ctx.RefreshGameData()
	utils.Sleep(500)

	ctx.WeaponBonusCache.Slot2AllClassBonus = calculateSlotBonus(ctx, 1, specificBonuses)
	utils.Sleep(200)

	if ctx.Data.ActiveWeaponSlot != 0 {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
		utils.Sleep(200)
	}

	optimalSlots := make(map[skill.ID]int)
	for skillID := range specificBonuses {
		slot0Total := ctx.WeaponBonusCache.Slot1AllClassBonus + specificBonuses[skillID]
		slot1Total := ctx.WeaponBonusCache.Slot2AllClassBonus + specificBonuses[skillID]

		if slot0Total >= slot1Total {
			optimalSlots[skillID] = 0
		} else {
			optimalSlots[skillID] = 1
		}
	}

	ctx.WeaponBonusCache.SkillSpecificBonuses = specificBonuses
	ctx.WeaponBonusCache.OptimalSkillSlots = optimalSlots
	ctx.WeaponBonusCache.IsValid = true

	ctx.Logger.Debug("Weapon bonus cache built",
		slog.Int("slot0_bonus", ctx.WeaponBonusCache.Slot1AllClassBonus),
		slog.Int("slot1_bonus", ctx.WeaponBonusCache.Slot2AllClassBonus),
		slog.Any("specific_bonuses", specificBonuses),
		slog.Any("optimal_slots", optimalSlots),
		slog.Bool("is_valid", true),
	)
}

func getSpecificSkillBonuses(ctx *context.Status, skillIDs []skill.ID) map[skill.ID]int {
	bonuses := make(map[skill.ID]int)
	for _, skillID := range skillIDs {
		if s, found := ctx.Data.PlayerUnit.Stats.FindStat(97, int(skillID)); found {
			bonuses[skillID] = s.Value
		}
	}
	return bonuses
}

func calculateSlotBonus(ctx *context.Status, slot int, specificBonuses map[skill.ID]int) int {
	total := 0

	allSkills := 0
	for layer := 0; layer < 10; layer++ {
		if s, found := ctx.Data.PlayerUnit.Stats.FindStat(stat.AllSkills, layer); found {
			allSkills += s.Value
		}
	}
	total += allSkills

	classLayer := int(ctx.Data.PlayerUnit.Class)
	if s, found := ctx.Data.PlayerUnit.Stats.FindStat(stat.AddClassSkills, classLayer); found {
		total += s.Value
	}

	for _, bonus := range specificBonuses {
		total += bonus
	}

	ctx.Logger.Debug("Slot Bonus Calculation",
		slog.Int("slot", slot),
		slog.Int("all_skills", allSkills),
		slog.Int("class_skills", total-allSkills),
		slog.Int("specific_skills", len(specificBonuses)),
	)

	return total
}
