package action

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
)

var uiStatButtonPosition = map[stat.ID]data.Position{
	stat.Strength:  {X: 240, Y: 210},
	stat.Dexterity: {X: 240, Y: 290},
	stat.Vitality:  {X: 240, Y: 380},
	stat.Energy:    {X: 240, Y: 430},
}

var uiSkillPagePosition = [3]data.Position{
	{X: 1100, Y: 140},
	{X: 1010, Y: 140},
	{X: 910, Y: 140},
}

var uiSkillRowPosition = [6]int{190, 250, 310, 365, 430, 490}
var uiSkillColumnPosition = [3]int{920, 1010, 1095}

var uiStatButtonPositionLegacy = map[stat.ID]data.Position{
	stat.Strength:  {X: 430, Y: 180},
	stat.Dexterity: {X: 430, Y: 250},
	stat.Vitality:  {X: 430, Y: 360},
	stat.Energy:    {X: 430, Y: 435},
}

var uiSkillPagePositionLegacy = [3]data.Position{
	{X: 970, Y: 510},
	{X: 970, Y: 390},
	{X: 970, Y: 260},
}

var uiSkillRowPositionLegacy = [6]int{110, 195, 275, 355, 440, 520}
var uiSkillColumnPositionLegacy = [3]int{690, 770, 855}

func EnsureStatPoints() error {
	// This function will allocate stat points to the character based on the settings in the character configuration file.

	ctx := context.Get()
	ctx.SetLastAction("EnsureStatPoints")

	ch, isLevelingChar := ctx.Char.(context.LevelingCharacter)
	findUnusedStats, _ := ctx.Data.PlayerUnit.FindStat(stat.StatPoints, 0)

	if isLevelingChar && findUnusedStats.Value > 0 {
		ctx.Logger.Info("Allocating stat points...")

		settingsStatPoints := ch.StatPoints() // complete list of statpoints to be leveled per character settings

		// This section sorts the settingsStatPoints map in the order set by the character config files. Go iterates over maps in a random order, so this is necessary to avoid random ordering.
		// Collect stats from the map
		sortedStats := make([]stat.ID, 0, len(settingsStatPoints))
		for st := range settingsStatPoints {
			sortedStats = append(sortedStats, st)
		}

		// Sort the stats
		sort.Slice(sortedStats, func(i, j int) bool {
			return sortedStats[i] < sortedStats[j]
		})

		unusedStats := findUnusedStats.Value // This has to be outside the for loops, so that is can be reduced for each stat that is allocated within the 2 for loops below, without resetting for each stat.

		// this runs through only the included stats from the character config files, in the order STR, ENG, DEX, VIT per d2go definitions
		for i := range sortedStats {
			targetPoints := settingsStatPoints[sortedStats[i]]
			currentPoints, _ := ctx.Data.PlayerUnit.FindStat(sortedStats[i], 0)

			if currentPoints.Value >= targetPoints {
				continue
			}

			var statsToAllocate = targetPoints - currentPoints.Value

			for unusedStats >= 1 && statsToAllocate >= 1 {

				// check if character menu is already open
				if !ctx.Data.OpenMenus.Character {
					ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.CharacterScreen)
				}

				// check if legacy or standard button co-ordinates should be used
				var statBtnPosition data.Position
				if ctx.Data.LegacyGraphics {
					statBtnPosition = uiStatButtonPositionLegacy[currentPoints.ID]
				} else {
					statBtnPosition = uiStatButtonPosition[currentPoints.ID]
				}

				utils.Sleep(100)
				ctx.HID.Click(game.LeftButton, statBtnPosition.X, statBtnPosition.Y)
				utils.Sleep(300)

				// reduce the remaining stats to allocate after clicking
				unusedStats = unusedStats - 1
				statsToAllocate = statsToAllocate - 1

			}
		}
	}

	return step.CloseAllMenus()
}

func EnsureSkillPoints() error {
	ctx := context.Get()
	ctx.SetLastAction("EnsureSkillPoints")
	unusedSkillPoints, _ := ctx.Data.PlayerUnit.FindStat(stat.SkillPoints, 0)

	ch, isLevelingChar := ctx.Char.(context.LevelingCharacter)
	if isLevelingChar && !ch.ShouldResetSkills() && unusedSkillPoints.Value > 0 { // only run if it's running a leveling script, is not going to reset skills, and has unused skillpoints
		ctx.Logger.Info("Allocating skill points...")

		ctx := context.Get()
		lvl, _ := ctx.Data.PlayerUnit.FindStat(stat.Level, 0)
		settingsSkillPoints := ch.SkillPoints() // this is a complete list of skills to be leveled per character settings

		// Extract the first X skillpoints from the list of skills to be leveled, where X is the character level
		var charLevel = lvl.Value
		var input []skill.ID = settingsSkillPoints

		firstX := input[:charLevel]
		result := make([]string, len(firstX))
		for i, v := range firstX {
			result[i] = fmt.Sprint(v)
		}

		// Count unique number of skills to be leveled, for this characters level
		counts := make(map[skill.ID]int)
		for _, skillID := range firstX {
			counts[skillID]++
		}

		// Create a sorted list of skill and skill level pairs i.e. "36 5, 37 1" (firebolt level 5, warmth level 1)
		var skills []skill.ID
		for sk := range counts {
			skills = append(skills, sk)
		}

		// this sorts the skills from highest to lowest skill ID. So better skills are skilled earlier (after the prerequisite skills)
		sort.Slice(skills, func(i, j int) bool { return skills[i] > skills[j] })

		var sortedOutput string // this is the output of the sorted list of skill and skill level pairs
		for _, skillID := range skills {
			sortedOutput += fmt.Sprintf("%d %d\n", skillID, counts[skillID]) // counts is the target skill level
		}

		// Level each unskilled skill in the order shown in settingsSkillPoints
		for _, skillID := range settingsSkillPoints {
			skillDesc, skFound := skill.Desc[skillID]
			if !skFound {
				ctx.Logger.Error("Skill not found for character", "skill", skillID)
				continue
			}

			var calcSkillPoints int = int(ctx.Data.PlayerUnit.Skills[skillID].Level) // Current char skill level. Converts this from uint to int for the if logic below
			if calcSkillPoints >= 1 {                                                // if the actual skill level is greater than 1, then skip this skill
				continue
			}

			if !ctx.Data.OpenMenus.SkillTree {
				ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SkillTree)
			}

			utils.Sleep(100)
			if ctx.Data.LegacyGraphics {
				ctx.HID.Click(game.LeftButton, uiSkillPagePositionLegacy[skillDesc.Page-1].X, uiSkillPagePositionLegacy[skillDesc.Page-1].Y)
			} else {
				ctx.HID.Click(game.LeftButton, uiSkillPagePosition[skillDesc.Page-1].X, uiSkillPagePosition[skillDesc.Page-1].Y)
			}
			utils.Sleep(200)
			if ctx.Data.LegacyGraphics {
				ctx.HID.Click(game.LeftButton, uiSkillColumnPositionLegacy[skillDesc.Column-1], uiSkillRowPositionLegacy[skillDesc.Row-1])
			} else {
				ctx.HID.Click(game.LeftButton, uiSkillColumnPosition[skillDesc.Column-1], uiSkillRowPosition[skillDesc.Row-1])
			}
			utils.Sleep(500)
		}

		// Level each skill to target skill level, highlest skill IDs first
		var iteration int = 0
		for _, skillID := range skills {
			sortedOutput += fmt.Sprintf("%d %d\n", skillID, counts[skillID])
			pair := strings.Split(sortedOutput, "\n")[iteration]  // this will return a string of the skill and skill level pair
			count, _ := strconv.Atoi(strings.Split(pair, " ")[1]) // this will return a string of skill level
			iteration = iteration + 1                             // this iterates the for loop to run through each unique skill.

			var calcSkillPoints int = int(ctx.Data.PlayerUnit.Skills[skillID].Level)  // Current char skill level. Converts this from uint to int for the if logic below
			unusedSkillPoints, _ := ctx.Data.PlayerUnit.FindStat(stat.SkillPoints, 0) // Unused skillpoints. This is the total number of skillpoints that can be allocated to the character.

			if calcSkillPoints < count && unusedSkillPoints.Value > 0 { // if the actual skill level is less than the target skill level, and there are unused skillpoints still to be spent
				min := int(math.Min(float64(count-calcSkillPoints), float64(unusedSkillPoints.Value))) // to level up the minimum of either unused skillpoints or the difference between the target skill level and the actual skill level

				skillDesc, skFound := skill.Desc[skillID]
				if !skFound {
					ctx.Logger.Error("Skill not found for character", "skill", skillID)
					return nil
				}

				if !ctx.Data.OpenMenus.SkillTree {
					ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SkillTree)
				}

				i := 1
				for i <= min { // levels up the skill until unused skillpoints are spent or the target skill level is reached
					utils.Sleep(100)
					if ctx.Data.LegacyGraphics {
						ctx.HID.Click(game.LeftButton, uiSkillPagePositionLegacy[skillDesc.Page-1].X, uiSkillPagePositionLegacy[skillDesc.Page-1].Y)
					} else {
						ctx.HID.Click(game.LeftButton, uiSkillPagePosition[skillDesc.Page-1].X, uiSkillPagePosition[skillDesc.Page-1].Y)
					}
					utils.Sleep(200)
					if ctx.Data.LegacyGraphics {
						ctx.HID.Click(game.LeftButton, uiSkillColumnPositionLegacy[skillDesc.Column-1], uiSkillRowPositionLegacy[skillDesc.Row-1])
					} else {
						ctx.HID.Click(game.LeftButton, uiSkillColumnPosition[skillDesc.Column-1], uiSkillRowPosition[skillDesc.Row-1])
					}
					utils.Sleep(500)
					i = i + 1
				}

			}
		}
	}

	return step.CloseAllMenus()
}

func UpdateQuestLog() error {
	ctx := context.Get()
	ctx.SetLastAction("UpdateQuestLog")

	if _, isLevelingChar := ctx.Char.(context.LevelingCharacter); !isLevelingChar {
		return nil
	}

	ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.QuestLog)
	utils.Sleep(1000)

	return step.CloseAllMenus()
}
func getAvailableSkillKB() []data.KeyBinding {
	availableSkillKB := make([]data.KeyBinding, 0)
	ctx := context.Get()
	ctx.SetLastAction("getAvailableSkillKB")

	for _, sb := range ctx.Data.KeyBindings.Skills {
		if sb.SkillID == -1 && (sb.Key1[0] != 0 && sb.Key1[0] != 255) || (sb.Key2[0] != 0 && sb.Key2[0] != 255) {
			availableSkillKB = append(availableSkillKB, sb.KeyBinding)
		}
	}

	return availableSkillKB
}

func EnsureSkillBindings() error {
	ctx := context.Get()
	ctx.SetLastAction("EnsureSkillBindings")

	char, isLevelingChar := ctx.Char.(context.LevelingCharacter)
	if !isLevelingChar {
		return nil
	}

	mainSkill, skillsToBind := char.SkillsToBind()
	skillsToBind = append(skillsToBind, skill.TomeOfTownPortal)
	notBoundSkills := make([]skill.ID, 0)
	for _, sk := range skillsToBind {
		if _, found := ctx.Data.KeyBindings.KeyBindingForSkill(sk); !found && ctx.Data.PlayerUnit.Skills[sk].Level > 0 {
			notBoundSkills = append(notBoundSkills, sk)
		}
	}

	if len(notBoundSkills) > 0 {
		ctx.HID.Click(game.LeftButton, ui.SecondarySkillButtonX, ui.SecondarySkillButtonY)
		utils.Sleep(300)
		ctx.HID.MovePointer(10, 10)
		utils.Sleep(300)

		availableKB := getAvailableSkillKB()

		for i, sk := range notBoundSkills {
			skillPosition, found := calculateSkillPositionInUI(false, sk)
			if !found {
				continue
			}

			ctx.HID.MovePointer(skillPosition.X, skillPosition.Y)
			utils.Sleep(100)
			ctx.HID.PressKeyBinding(availableKB[i])
			utils.Sleep(300)
		}

	}

	if ctx.Data.PlayerUnit.LeftSkill != mainSkill {
		ctx.HID.Click(game.LeftButton, ui.MainSkillButtonX, ui.MainSkillButtonY)
		utils.Sleep(300)
		ctx.HID.MovePointer(10, 10)
		utils.Sleep(300)

		skillPosition, found := calculateSkillPositionInUI(true, mainSkill)
		if found {
			ctx.HID.MovePointer(skillPosition.X, skillPosition.Y)
			utils.Sleep(100)
			ctx.HID.Click(game.LeftButton, skillPosition.X, skillPosition.Y)
			utils.Sleep(300)
		}
	}

	return nil
}

func calculateSkillPositionInUI(mainSkill bool, skillID skill.ID) (data.Position, bool) {
	d := context.Get().Data

	var scrolls = []skill.ID{
		skill.TomeOfTownPortal, skill.ScrollOfTownPortal, skill.TomeOfIdentify, skill.ScrollOfIdentify,
	}

	if _, found := d.PlayerUnit.Skills[skillID]; !found {
		return data.Position{}, false
	}

	targetSkill := skill.Skills[skillID]
	descs := make(map[skill.ID]skill.Skill)
	row := 0
	totalRows := make([]int, 0)
	column := 0
	skillsWithCharges := 0
	for skID, points := range d.PlayerUnit.Skills {
		sk := skill.Skills[skID]
		// Skip skills that can not be bind
		if sk.Desc().ListRow < 0 {
			continue
		}

		// Skip skills that can not be bind to current mouse button
		if (mainSkill == true && !sk.LeftSkill) || (mainSkill == false && !sk.RightSkill) {
			continue
		}

		if points.Charges > 0 {
			skillsWithCharges++
			continue
		}

		if slices.Contains(scrolls, sk.ID) {
			continue
		}
		descs[skID] = sk

		if skID != targetSkill.ID && sk.Desc().Page == targetSkill.Desc().Page {
			if sk.Desc().ListRow > targetSkill.Desc().ListRow {
				column++
			} else if sk.Desc().ListRow == targetSkill.Desc().ListRow && sk.Desc().Column > targetSkill.Desc().Column {
				column++
			}
		}

		totalRows = append(totalRows, sk.Desc().ListRow)
		if row == targetSkill.Desc().ListRow {
			continue
		}

		row++
	}

	slices.Sort(totalRows)
	totalRows = slices.Compact(totalRows)

	// If we don't have any skill of a specific tree, the entire row gets one line down
	for i, currentRow := range totalRows {
		if currentRow == row {
			row = i
			break
		}
	}

	// Scrolls and charges are not in the same list
	if slices.Contains(scrolls, skillID) {
		column = skillsWithCharges
		row = len(totalRows)
		for _, skID := range scrolls {
			if d.PlayerUnit.Skills[skID].Quantity > 0 {
				if skID == skillID {
					break
				}
				column++
			}
		}
	}

	skillOffsetX := ui.MainSkillListFirstSkillX - (ui.SkillListSkillOffset * column)
	if !mainSkill {
		skillOffsetX = ui.SecondarySkillListFirstSkillX + (ui.SkillListSkillOffset * column)
	}

	return data.Position{
		X: skillOffsetX,
		Y: ui.SkillListFirstSkillY - ui.SkillListSkillOffset*row,
	}, true
}

func HireMerc() error {
	ctx := context.Get()
	ctx.SetLastAction("HireMerc")

	_, isLevelingChar := ctx.Char.(context.LevelingCharacter)
	if isLevelingChar && ctx.CharacterCfg.Character.UseMerc {
		// Hire the merc if we don't have one, we have enough gold, and we are in act 2. We assume that ReviveMerc was called before this.
		if ctx.CharacterCfg.Game.Difficulty == difficulty.Normal && ctx.Data.MercHPPercent() <= 0 && ctx.Data.PlayerUnit.TotalPlayerGold() > 30000 && ctx.Data.PlayerUnit.Area == area.LutGholein {
			ctx.Logger.Info("Hiring merc...")
			// TODO: Hire Holy Freeze merc if available, if not, hire Defiance merc.
			err := InteractNPC(town.GetTownByArea(ctx.Data.PlayerUnit.Area).MercContractorNPC())
			if err != nil {
				return err
			}
			ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)
			utils.Sleep(2000)
			ctx.HID.Click(game.LeftButton, ui.FirstMercFromContractorListX, ui.FirstMercFromContractorListY)
			utils.Sleep(500)
			ctx.HID.Click(game.LeftButton, ui.FirstMercFromContractorListX, ui.FirstMercFromContractorListY)
		}
	}

	return nil
}

func ResetStats() error {
	ctx := context.Get()
	ctx.SetLastAction("ResetStats")

	ch, isLevelingChar := ctx.Char.(context.LevelingCharacter)
	if isLevelingChar && ch.ShouldResetSkills() {
		currentArea := ctx.Data.PlayerUnit.Area
		if ctx.Data.PlayerUnit.Area != area.RogueEncampment {
			err := WayPoint(area.RogueEncampment)
			if err != nil {
				return err
			}
		}
		InteractNPC(npc.Akara)
		ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_DOWN, win.VK_RETURN)
		utils.Sleep(1000)
		ctx.HID.KeySequence(win.VK_HOME, win.VK_RETURN)

		if currentArea != area.RogueEncampment {
			return WayPoint(currentArea)
		}
	}

	return nil
}

func WaitForAllMembersWhenLeveling() error {
	ctx := context.Get()
	ctx.SetLastAction("WaitForAllMembersWhenLeveling")

	for {
		_, isLeveling := ctx.Char.(context.LevelingCharacter)
		if ctx.CharacterCfg.Companion.Leader && !ctx.Data.PlayerUnit.Area.IsTown() && isLeveling {
			allMembersAreaCloseToMe := true
			for _, member := range ctx.Data.Roster {
				if member.Name != ctx.Data.PlayerUnit.Name && ctx.PathFinder.DistanceFromMe(member.Position) > 20 {
					allMembersAreaCloseToMe = false
				}
			}

			if allMembersAreaCloseToMe {
				return nil
			}

			ClearAreaAroundPlayer(5, data.MonsterAnyFilter())
		} else {
			return nil
		}
	}
}
