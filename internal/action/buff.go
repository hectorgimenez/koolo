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

			}
			castSkills(ctx, slot0Pre)

			if len(slot1Pre) > 0 && ctx.Data.ActiveWeaponSlot != 1 {
				ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
				utils.Sleep(200)

			}
			castSkills(ctx, slot1Pre)
		} else {

			castSkills(ctx, preCTASkills)
		}
	}

	// If i exclude the berserker class from buffCTA, then he could still have a CTA equipped and use it for BuffSkills
	if ctx.CharacterCfg.Character.Class != "berserker" {
		buffCTA()
	}
	postCTASkills := ctx.Char.BuffSkills()
	if len(postCTASkills) > 0 {
		ctx.Logger.Debug("Post CTA Buffing...")

		if ctx.WeaponBonusCache.IsValid {

			slot0Post, slot1Post := groupSkillsBySlot(ctx, postCTASkills)

			if len(slot0Post) > 0 && ctx.Data.ActiveWeaponSlot != 0 {
				utils.Sleep(300)
				ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)

			}

			castSkills(ctx, slot0Post)
			if len(slot1Post) > 0 && ctx.Data.ActiveWeaponSlot != 1 {
				utils.Sleep(300)
				ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)

			}

			castSkills(ctx, slot1Post)

		} else {
			castSkills(ctx, postCTASkills)
		}

	}
	step.SwapToMainWeapon()
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
				ctx.Logger.Info("Key binding not found", slog.String("skill", skill.SkillNames[skillID]))
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

		//Since we already swap to mainweapon at the end of Buff() we don't need to swap back here
		//utils.Sleep(500)
		//step.SwapToMainWeapon()
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

	skills := append(ctx.Char.PreCTABuffSkills(), ctx.Char.BuffSkills()...)

	specificBonusesSlot0 := getSpecificSkillBonuses(ctx, skills)
	slot0Total := calculateSlotBonus(ctx)
	ctx.WeaponBonusCache.Slot0AllClassBonus = slot0Total
	utils.Sleep(200)

	ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
	ctx.RefreshGameData()
	utils.Sleep(500)

	specificBonusesSlot1 := getSpecificSkillBonuses(ctx, skills)
	slot1Total := calculateSlotBonus(ctx)
	ctx.WeaponBonusCache.Slot1AllClassBonus = slot1Total
	utils.Sleep(200)

	if ctx.Data.ActiveWeaponSlot != 0 {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SwapWeapons)
		utils.Sleep(200)
	}

	optimalSlots := make(map[skill.ID]int)
	unassignedSlots := make([]skill.ID, 0)
	slot0Count := 0
	slot1Count := 0
	for _, skillID := range skills {
		slot0Bonus := ctx.WeaponBonusCache.Slot0AllClassBonus
		slot1Bonus := ctx.WeaponBonusCache.Slot1AllClassBonus

		if bonus, ok := specificBonusesSlot0[skillID]; ok {
			slot0Bonus += bonus

		}
		if bonus, ok := specificBonusesSlot1[skillID]; ok {
			slot1Bonus += bonus
		}

		if slot0Bonus > slot1Bonus {
			optimalSlots[skillID] = 0
			slot0Count++
		} else if slot0Bonus < slot1Bonus {
			optimalSlots[skillID] = 1
			slot1Count++
		} else {
			unassignedSlots = append(unassignedSlots, skillID)
		}
	}

	if slot0Count > slot1Count {
		for _, skillID := range unassignedSlots {
			optimalSlots[skillID] = 0
		}
	} else {
		for _, skillID := range unassignedSlots {
			optimalSlots[skillID] = 1
		}
	}

	ctx.WeaponBonusCache.OptimalSkillSlots = optimalSlots
	ctx.WeaponBonusCache.IsValid = true

	ctx.Logger.Debug("Weapon bonus cache built",
		slog.Int("slot0_bonus", slot0Total),
		slog.Int("slot1_bonus", slot1Total),
		slog.Any("specific_bonuses0", specificBonusesSlot0),
		slog.Any("specific_bonuses1", specificBonusesSlot1),
		slog.Any("optimal_slots", optimalSlots),
		slog.Bool("is_valid", true),
	)
}

func getSpecificSkillBonuses(ctx *context.Status, skillIDs []skill.ID) map[skill.ID]int {
	bonuses := make(map[skill.ID]int)

	for _, skillID := range skillIDs {
		bonus := 0

		if s, found := ctx.Data.PlayerUnit.Stats.FindStat(stat.SingleSkill, int(skillID)); found {
			bonus += s.Value

		}

		if s, found := ctx.Data.PlayerUnit.Stats.FindStat(stat.NonClassSkill, int(skillID)); found {
			bonus += s.Value

		}

		if tabID, isMapped := SkillToTabs[skillID]; isMapped {
			if s, found := ctx.Data.PlayerUnit.Stats.FindStat(stat.AddSkillTab, tabID); found {
				bonus += s.Value
			}
		}

		if bonus > 0 {
			bonuses[skillID] = bonus
			ctx.Logger.Debug("Skill bonus total",
				slog.Int("slot", ctx.Data.ActiveWeaponSlot),
				slog.String("skillname", skill.SkillNames[skillID]),
				slog.Int("skill_id", int(skillID)),
				slog.Int("total", bonus))
		}
	}

	return bonuses
}

func calculateSlotBonus(ctx *context.Status) int {
	total := 0

	allSkills := 0
	for layer := 0; layer < 10; layer++ {
		if s, found := ctx.Data.PlayerUnit.Stats.FindStat(stat.AllSkills, layer); found {
			allSkills += s.Value
		}
	}
	total += allSkills

	classSkills := 0
	if ctx.Data.PlayerUnit.Class > 0 {
		classLayer := int(ctx.Data.PlayerUnit.Class)
		if s, found := ctx.Data.PlayerUnit.Stats.FindStat(stat.AddClassSkills, classLayer); found {
			classSkills = s.Value
		}
	}
	total += classSkills

	return total
}

// For this to work, right now, we need this List, because Druid Skills and Assassin Skills are missing in "github.com/hectorgimenez/d2go/pkg/data/skill/skilldesc.go"
var SkillToTabs = map[skill.ID]int{
	skill.AmplifyDamage:     16,
	skill.ArcticBlast:       42,
	skill.Armageddon:        42,
	skill.Attract:           16,
	skill.Avoid:             1,
	skill.AxeMastery:        33,
	skill.Bash:              32,
	skill.BattleCommand:     34,
	skill.BattleCry:         34,
	skill.BattleOrders:      34,
	skill.Berserk:           32,
	skill.BladeFury:         48,
	skill.BladeSentinel:     48,
	skill.BladeShield:       48,
	skill.BladesOfIce:       50,
	skill.Blaze:             8,
	skill.BlessedAim:        25,
	skill.BlessedHammer:     32,
	skill.Blizzard:          10,
	skill.BloodGolem:        18,
	skill.BoneArmor:         17,
	skill.BonePrison:        17,
	skill.BoneSpear:         17,
	skill.BoneSpirit:        17,
	skill.BoneWall:          17,
	skill.BurstOfSpeed:      49,
	skill.CarrionVine:       40,
	skill.ChainLightning:    9,
	skill.Charge:            32,
	skill.ChargedBolt:       9,
	skill.ChargedBoltSentry: 48,
	skill.ChargedStrike:     2,
	skill.ChillingArmor:     10,
	skill.ClawMastery:       49,
	skill.ClawsOfThunder:    50,
	skill.ClayGolem:         18,
	skill.Cleansing:         26,
	skill.CloakOfShadows:    49,
	skill.CobraStrike:       50,
	skill.ColdArrow:         0,
	skill.ColdMastery:       10,
	skill.Concentrate:       32,
	skill.Concentration:     25,
	skill.Confuse:           16,
	skill.Conversion:        32,
	skill.Conviction:        25,
	skill.CorpseExplosion:   17,
	skill.CriticalStrike:    1,
	skill.CycloneArmor:      42,
	skill.DeathSentry:       48,
	skill.Decoy:             1,
	skill.Decrepify:         16,
	skill.Defiance:          26,
	skill.DimVision:         16,
	skill.Dodge:             1,
	skill.DoubleSwing:       32,
	skill.DoubleThrow:       32,
	skill.DragonClaw:        50,
	skill.DragonFlight:      50,
	skill.DragonTail:        50,
	skill.DragonTalon:       50,
	skill.Enchant:           8,
	skill.EnergyShield:      9,
	skill.Evade:             1,
	skill.ExplodingArrow:    0,
	skill.Fade:              49,
	skill.Fanaticism:        25,
	skill.Fend:              2,
	skill.FeralRage:         41,
	skill.FindItem:          34,
	skill.FindPotion:        34,
	skill.FireArrow:         0,
	skill.FireBlast:         48,
	skill.FireBolt:          8,
	skill.FireClaws:         41,
	skill.FireGolem:         18,
	skill.FireMastery:       8,
	skill.FireWall:          8,
	skill.FireBall:          8,
	skill.Firestorm:         42,
	skill.Fissure:           42,
	skill.FistOfTheHeavens:  32,
	skill.FistsOfFire:       50,
	skill.FreezingArrow:     0,
	skill.Frenzy:            32,
	skill.FrostNova:         10,
	skill.FrozenArmor:       10,
	skill.FrozenOrb:         10,
	skill.Fury:              41,
	skill.GlacialSpike:      10,
	skill.GolemMastery:      18,
	skill.GrimWard:          34,
	skill.GuidedArrow:       0,
	skill.HeartOfWolverine:  40,
	skill.HolyBolt:          32,
	skill.HolyFire:          25,
	skill.HolyFreeze:        25,
	skill.HolyShield:        32,
	skill.HolyShock:         25,
	skill.Howl:              34,
	skill.Hunger:            41,
	skill.Hurricane:         42,
	skill.Hydra:             8,
	skill.IceArrow:          0,
	skill.IceBlast:          10,
	skill.IceBolt:           10,
	skill.ImmolationArrow:   0,
	skill.Impale:            2,
	skill.IncreasedSpeed:    33,
	skill.IncreasedStamina:  33,
	skill.Inferno:           8,
	skill.InnerSight:        1,
	skill.IronGolem:         18,
	skill.IronMaiden:        16,
	skill.IronSkin:          33,
	skill.Jab:               2,
	skill.Leap:              32,
	skill.LeapAttack:        32,
	skill.LifeTap:           16,
	skill.Lightning:         9,
	skill.LightningBolt:     2,
	skill.LightningFury:     2,
	skill.LightningMastery:  9,
	skill.LightningSentry:   48,
	skill.LightningStrike:   2,
	skill.LowerResist:       16,
	skill.Lycanthropy:       41,
	skill.MaceMastery:       33,
	skill.MagicArrow:        0,
	skill.Maul:              41,
	skill.Meditation:        26,
	skill.Meteor:            8,
	skill.Might:             25,
	skill.MindBlast:         49,
	skill.MoltenBoulder:     42,
	skill.MultipleShot:      0,
	skill.NaturalResistance: 33,
	skill.Nova:              9,
	skill.OakSage:           40,
	skill.Penetrate:         1,
	skill.PhoenixStrike:     50,
	skill.Pierce:            1,
	skill.PlagueJavelin:     2,
	skill.PoisonCreeper:     40,
	skill.PoisonDagger:      17,
	skill.PoisonExplosion:   17,
	skill.PoisonJavelin:     2,
	skill.PoisonNova:        17,
	skill.PolearmMastery:    33,
	skill.PowerStrike:       2,
	skill.Prayer:            26,
	skill.PsychicHammer:     49,
	skill.Rabies:            41,
	skill.RaiseSkeleton:     18,
	skill.Raven:             40,
	skill.Redemption:        26,
	skill.ResistCold:        26,
	skill.ResistFire:        26,
	skill.ResistLightning:   26,
	skill.Revive:            18,
	skill.Sacrifice:         32,
	skill.Salvation:         26,
	skill.Sanctuary:         25,
	skill.ShadowMaster:      49,
	skill.ShadowWarrior:     49,
	skill.ShiverArmor:       10,
	skill.ShockWave:         41,
	skill.ShockWeb:          48,
	skill.Shout:             34,
	skill.RaiseSkeletalMage: 18,
	skill.SkeletonMastery:   18,
	skill.SlowMissiles:      1,
	skill.Smite:             32,
	skill.SolarCreeper:      40,
	skill.SpearMastery:      33,
	skill.SpiritOfBarbs:     40,
	skill.StaticField:       9,
	skill.Strafe:            0,
	skill.Stun:              32,
	skill.SummonDireWolf:    40,
	skill.SummonGrizzly:     40,
	skill.SummonResist:      18,
	skill.SummonSpiritWolf:  40,
	skill.Taunt:             34,
	skill.Teeth:             17,
	skill.Telekinesis:       9,
	skill.Teleport:          9,
	skill.Terror:            16,
	skill.Thorns:            25,
	skill.ThrowingMastery:   33,
	skill.ThunderStorm:      9,
	skill.TigerStrike:       50,
	skill.Tornado:           42,
	skill.Twister:           42,
	skill.Valkyrie:          1,
	skill.Vengeance:         32,
	skill.Venom:             49,
	skill.Vigor:             26,
	skill.Volcano:           42,
	skill.WakeOfFire:        48,
	skill.WakeOfInferno:     48,
	skill.WarCry:            34,
	skill.Warmth:            8,
	skill.Weaken:            16,
	skill.WeaponBlock:       49,
	skill.Werebear:          41,
	skill.Werewolf:          41,
	skill.Whirlwind:         32,
	skill.Zeal:              32,
}
