package action

import (
	"github.com/hectorgimenez/d2go/pkg/data/skill"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/context"
)

const (
	BaseScore     = 1.0
	BeltBaseSlots = 4
)

var (
	skillWeights = map[stat.ID]float64{
		stat.AllSkills:      200.0,
		stat.AddClassSkills: 175.0,
		stat.AddSkillTab:    125.0,
	}

	resistWeightsMain = map[stat.ID]float64{
		stat.FireResist:      3.0,
		stat.ColdResist:      2.0,
		stat.LightningResist: 3.0,
		stat.PoisonResist:    1.0,
	}

	resistWeightsOther = map[stat.ID]float64{
		stat.MaxFireResist:          8.0,
		stat.MaxLightningResist:     8.0,
		stat.MaxColdResist:          6.0,
		stat.MaxPoisonResist:        4.0,
		stat.AbsorbFire:             2.0,
		stat.AbsorbLightning:        2.0,
		stat.AbsorbMagic:            2.0,
		stat.AbsorbCold:             2.0,
		stat.AbsorbFirePercent:      4.0,
		stat.AbsorbLightningPercent: 4.0,
		stat.AbsorbMagicPercent:     4.0,
		stat.AbsorbColdPercent:      4.0,
		stat.DamageReduced:          2.0,
		stat.DamagePercent:          3.0,
		stat.MagicDamageReduction:   2.0,
		stat.MagicResist:            2.0,
	}

	generalWeights = map[stat.ID]float64{
		stat.CannotBeFrozen:    25.0,
		stat.FasterHitRecovery: 3.0,
		stat.FasterRunWalk:     2.0,
		stat.FasterBlockRate:   2.0,
		stat.FasterCastRate:    4.0,
		stat.ChanceToBlock:     2.5,
		stat.MagicFind:         1.0,
		stat.GoldFind:          0.1,
		stat.Defense:           0.05,
		stat.ManaRecovery:      2.0,
		stat.Strength:          1.0,
		stat.Dexterity:         1.0,
		stat.Vitality:          1.5,
		stat.Energy:            0.5,
		stat.MaxLife:           0.5,
		stat.MaxMana:           0.25,
		stat.ReplenishQuantity: 2.0,
		stat.ReplenishLife:     2.0,
		stat.LifePerLevel:      3.0,
		stat.ManaPerLevel:      2.0,
	}

	mercWeights = map[stat.ID]float64{
		stat.IncreasedAttackSpeed:   3.5,
		stat.MinDamage:              3.0,
		stat.MaxDamage:              3.0,
		stat.TwoHandedMinDamage:     3.0,
		stat.TwoHandedMaxDamage:     3.0,
		stat.AttackRating:           0.1,
		stat.CrushingBlow:           3.0,
		stat.OpenWounds:             3.0,
		stat.LifeSteal:              8.0,
		stat.ReplenishLife:          2.0,
		stat.FasterHitRecovery:      3.0,
		stat.Defense:                0.05,
		stat.Strength:               1.5,
		stat.Dexterity:              1.5,
		stat.FireResist:             2.0,
		stat.ColdResist:             1.5,
		stat.LightningResist:        2.0,
		stat.PoisonResist:           1.0,
		stat.DamageReduced:          2.0,
		stat.MagicResist:            3.0,
		stat.AbsorbFirePercent:      2.7,
		stat.AbsorbColdPercent:      2.7,
		stat.AbsorbLightningPercent: 2.7,
		stat.AbsorbMagicPercent:     2.7,
	}

	beltSizes = map[string]int{
		"lbl": 2,
		"vbl": 2,
		"mbl": 3,
		"tbl": 3,
	}
)

type mercCTCWeights struct {
	StatID stat.ID
	Weight float64
	Layer  int
}

type ResistStats struct {
	Fire      int
	Cold      int
	Lightning int
	Poison    int
}

// Example usage:
var mercCTCWeight = []mercCTCWeights{
	{StatID: stat.SkillOnAttack, Weight: 5.0, Layer: 4227},     // Amp Damage
	{StatID: stat.SkillOnAttack, Weight: 10.0, Layer: 5572},    // Decrepify
	{StatID: stat.SkillOnHit, Weight: 3.0, Layer: 4225},        // Amp Damage
	{StatID: stat.SkillOnHit, Weight: 8.0, Layer: 5572},        // Decrepify
	{StatID: stat.SkillOnGetHit, Weight: 1000.0, Layer: 17103}, // Fade
	{StatID: stat.Aura, Weight: 1000.0, Layer: 123},            // Infinity
	{StatID: stat.Aura, Weight: 100.0, Layer: 120},             // Insight
}

// PlayerScore calculates overall item tier score
func PlayerScore(itm data.Item) float64 {

	generalScore := calculateGeneralScore(itm)
	resistScore := calculateResistScore(itm)
	skillScore := calculateSkillScore(itm)

	totalScore := BaseScore + generalScore + resistScore + skillScore

	return totalScore
}
func calculateGeneralScore(itm data.Item) float64 {

	score := BaseScore
	// Handle Cannot Be Frozen
	//if !ctx.Data.CanTeleport() && itm.FindStat(stat.CannotbeFrozen, 0) {
	//	if <add logic to check if another item has CBF> {
	//		score += GeneralWeights[stat.CannotbeFrozen]
	//	}
	//}

	if itm.Desc().Type == "belt" {
		score += calculateBeltScore(itm)
	}

	// Handle sockets - this might be a bad idea becauase we won't properly use the sockets
	// May also double count if the sockets are filled
	if !itm.IsRuneword {
		score += float64(itm.Desc().MaxSockets * 10)
	}

	score += calculatePerLevelStats(itm)
	score += calculateBaseStats(itm)

	return score
}

// Belt-specific scoring so we don't lose belt slots
func calculateBeltScore(itm data.Item) float64 {
	beltSize := getBeltSize(itm)
	currentSize := getCurrentBeltSize()

	// Slots matter more than stats, this should never downgrade a belt
	if currentSize > beltSize {
		return float64(-1000)
	}
	return float64(beltSize * BeltBaseSlots * 2)
}

func getBeltSize(itm data.Item) int {
	if size := beltSizes[itm.Desc().Code]; size > 0 {
		return size
	}
	return BeltBaseSlots
}

func getCurrentBeltSize() int {
	ctx := context.Get()
	for _, item := range ctx.Data.Inventory.ByLocation(item.LocationEquipped) {
		if item.Desc().Type == "belt" {
			return beltSizes[item.Desc().Code]
		}
	}
	return 0
}

func calculatePerLevelStats(itm data.Item) float64 {
	ctx := context.Get()
	charLevel, _ := ctx.Data.PlayerUnit.FindStat(stat.Level, 0)

	lifePerlvl, _ := itm.FindStat(stat.LifePerLevel, 0)
	manaPerlvl, _ := itm.FindStat(stat.ManaPerLevel, 0)

	return (float64(lifePerlvl.Value/2048)*float64(charLevel.Value))*generalWeights[stat.LifePerLevel] +
		(float64(manaPerlvl.Value/2048)*float64(charLevel.Value))*generalWeights[stat.ManaPerLevel]
}

// calculateBaseStats evaluates common item stats
func calculateBaseStats(itm data.Item) float64 {
	score := 0.0

	for statID, weight := range generalWeights {
		if statData, found := itm.FindStat(statID, 0); found {
			statScore := float64(statData.Value) * weight
			score += statScore
		}
	}

	return score
}

// Resists

// calculateResistScore evaluates item resistance values and returns a weighted score
func calculateResistScore(itm data.Item) float64 {
	// Get new item resists
	newResists := getItemMainResists(itm)

	// only enter next block if we have a new item with resists
	if newResists.Fire == 0 && newResists.Cold == 0 && newResists.Lightning == 0 && newResists.Poison == 0 {
		return 0.0
	}

	// get item resists stats from olditem equipped on body location
	oldResists := getEquippedResists(itm)

	baseResists := getBaseResists(oldResists)
	// subtract olditem resists from current total resists
	effectiveResists := calculateEffectiveResists(newResists, baseResists)

	mainScore := calculateMainResistScore(effectiveResists)
	otherScore := calculateOtherResistScore(itm)

	return mainScore + otherScore
}

func getItemMainResists(itm data.Item) ResistStats {
	fr, _ := itm.FindStat(stat.FireResist, 0)
	cr, _ := itm.FindStat(stat.ColdResist, 0)
	lr, _ := itm.FindStat(stat.LightningResist, 0)
	pr, _ := itm.FindStat(stat.PoisonResist, 0)

	return ResistStats{
		Fire:      fr.Value,
		Cold:      cr.Value,
		Lightning: lr.Value,
		Poison:    pr.Value,
	}
}

func getEquippedResists(newItem data.Item) ResistStats {
	ctx := context.Get()
	var resists ResistStats

	for _, equippedItem := range ctx.Data.Inventory.ByLocation(item.LocationEquipped) {
		if equippedItem.Desc().Type[0] == newItem.Desc().Type[0] {
			fr, _ := equippedItem.FindStat(stat.FireResist, 0)
			resists.Fire = fr.Value
			cr, _ := equippedItem.FindStat(stat.ColdResist, 0)
			resists.Cold = cr.Value
			lr, _ := equippedItem.FindStat(stat.LightningResist, 0)
			resists.Lightning = lr.Value
			pr, _ := equippedItem.FindStat(stat.PoisonResist, 0)
			resists.Poison = pr.Value
			break
		}
	}
	return resists
}

func getBaseResists(equipped ResistStats) ResistStats {
	ctx := context.Get()

	fr, _ := ctx.Data.PlayerUnit.BaseStats.FindStat(stat.FireResist, 0)
	cr, _ := ctx.Data.PlayerUnit.BaseStats.FindStat(stat.ColdResist, 0)
	lr, _ := ctx.Data.PlayerUnit.BaseStats.FindStat(stat.LightningResist, 0)
	pr, _ := ctx.Data.PlayerUnit.BaseStats.FindStat(stat.PoisonResist, 0)

	return ResistStats{
		Fire:      fr.Value - equipped.Fire,
		Cold:      cr.Value - equipped.Cold,
		Lightning: lr.Value - equipped.Lightning,
		Poison:    pr.Value - equipped.Poison,
	}
}

func calculateEffectiveResists(new, base ResistStats) ResistStats {
	const maxResist = 75

	return ResistStats{
		Fire:      min(new.Fire, max(maxResist-base.Fire, 0)),
		Cold:      min(new.Cold, max(maxResist-base.Cold, 0)),
		Lightning: min(new.Lightning, max(maxResist-base.Lightning, 0)),
		Poison:    min(new.Poison, max(maxResist-base.Poison, 0)),
	}
}

func calculateMainResistScore(resists ResistStats) float64 {
	return float64(resists.Fire)*resistWeightsMain[stat.FireResist] +
		float64(resists.Cold)*resistWeightsMain[stat.ColdResist] +
		float64(resists.Lightning)*resistWeightsMain[stat.LightningResist] +
		float64(resists.Poison)*resistWeightsMain[stat.PoisonResist]
}

func calculateOtherResistScore(itm data.Item) float64 {
	var score float64

	for statID, weight := range resistWeightsOther {
		if statData, found := itm.FindStat(statID, 0); found {
			score += float64(statData.Value) * weight
		}
	}

	return score
}

// Skill calcs

func calculateSkillScore(itm data.Item) float64 {
	ctx := context.Get()
	score := 0.0

	if statData, found := itm.FindStat(stat.AllSkills, 0); found {
		allSkillScore := float64(statData.Value) * skillWeights[statData.ID]
		score += allSkillScore
	}

	if classSkillsStat, found := itm.FindStat(stat.AddClassSkills, int(ctx.Data.PlayerUnit.Class)); found {
		classSkillScore := float64(classSkillsStat.Value) * skillWeights[classSkillsStat.ID]
		score += classSkillScore
	}

	tabskill := int(ctx.Data.PlayerUnit.Class)*8 + (getMaxSkillTabPage() - 1)
	if tabSkillsStat, found := itm.FindStat(stat.AddSkillTab, tabskill); found {
		tabSkillScore := float64(tabSkillsStat.Value) * skillWeights[tabSkillsStat.ID]
		score += tabSkillScore
	}

	usedSkills := make([]skill.ID, 0)

	//Let's ignore 1 point wonders unless we're above level 2
	for sk, pts := range ctx.Data.PlayerUnit.Skills {
		if pts.Level > 1 {
			usedSkills = append(usedSkills, sk)
		} else if lvl, _ := ctx.Data.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 3 {
			usedSkills = append(usedSkills, sk)
		}
	}

	// TODO Multiply the +x part of the skill bonus instead of returning a flat value
	for _, usedSkill := range usedSkills {
		if _, found := itm.FindStat(stat.SingleSkill, int(usedSkill)); found {
			score += 40
		}
	}

	return score
}

// MercScore calculates mercenary-specific item score
func MercScore(itm data.Item) float64 {
	score := BaseScore + sumElementalDamage(itm)*2.0

	// Add base stat scores
	for statID, weight := range mercWeights {
		if statData, found := itm.FindStat(statID, 0); found {
			mercStatScore := float64(statData.Value) * weight
			score += mercStatScore
		}
	}

	// Add cast-on-trigger scores
	for _, ctc := range mercCTCWeight {
		if _, found := itm.FindStat(ctc.StatID, ctc.Layer); found {
			score += ctc.Weight
		}
	}

	return score
}

// Helper functions
func sumElementalDamage(itm data.Item) float64 {
	return sumDamageType(itm, stat.FireMinDamage, stat.FireMaxDamage) +
		sumDamageType(itm, stat.LightningMinDamage, stat.LightningMaxDamage) +
		sumDamageType(itm, stat.ColdMinDamage, stat.ColdMaxDamage) +
		sumDamageType(itm, stat.MagicMinDamage, stat.MagicMaxDamage) +
		calculatePoisonDamage(itm)
}

func sumDamageType(itm data.Item, minStat, maxStat stat.ID) float64 {
	min, _ := itm.FindStat(minStat, 0)
	max, _ := itm.FindStat(maxStat, 0)
	return float64(min.Value + max.Value)
}

func calculatePoisonDamage(itm data.Item) float64 {
	poisonMin, _ := itm.FindStat(stat.PoisonMinDamage, 0)
	return float64(poisonMin.Value) * 125.0 / 256.0
}

func getMaxSkillTabPage() int {
	ctx := context.Get()

	tabCounts := make(map[int]int)
	maxCount := 0
	maxPage := 0
	for pskill, pts := range ctx.Data.PlayerUnit.Skills {
		if page := pskill.Desc().Page; page > 0 {
			tabCounts[page] += int(pts.Level)
			if tabCounts[page] > maxCount {
				maxCount = tabCounts[page]
				maxPage = page
			}
		}
	}

	return maxPage
}
