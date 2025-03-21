package action

import (
	"fmt"
	"slices"

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
	"github.com/hectorgimenez/d2go/pkg/data/item"
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
	ctx := context.Get()
	char, isLevelingChar := ctx.Char.(context.LevelingCharacter)
	if !isLevelingChar {
		return nil
	}

	statPoints, hasUnusedPoints := ctx.Data.PlayerUnit.FindStat(stat.StatPoints, 0)
	if !hasUnusedPoints || statPoints.Value == 0 {
		return nil
	}

	remainingPoints := statPoints.Value
	allocations := char.StatPoints()
	for _, allocation := range allocations {
		if statPoints.Value == 0 {
			break
		}

		currentValue, _ := ctx.Data.PlayerUnit.BaseStats.FindStat(allocation.Stat, 0)
		if currentValue.Value >= allocation.Points {
			continue
		}

		// Calculate how many points we can actually spend
		pointsToSpend := min(allocation.Points-currentValue.Value, remainingPoints)
		for i := 0; i < pointsToSpend; i++ {

			if !spendStatPoint(allocation.Stat) {
				ctx.Logger.Error(fmt.Sprintf("Failed to spend point in %v", allocation.Stat))
				continue
			}

			remainingPoints--

			updatedValue, _ := ctx.Data.PlayerUnit.BaseStats.FindStat(allocation.Stat, 0)
			if updatedValue.Value >= allocation.Points {
				ctx.Logger.Debug(fmt.Sprintf("Increased %v to target %d (%d total points remaining)",
					allocation.Stat, allocation.Points, remainingPoints))
			}
		}
	}
	return step.CloseAllMenus()

}

func EnsureSkillPoints() error {
	ctx := context.Get()

	char, isLevelingChar := ctx.Char.(context.LevelingCharacter)
	skillPoints, hasUnusedPoints := ctx.Data.PlayerUnit.FindStat(stat.SkillPoints, 0)

	if !isLevelingChar || !hasUnusedPoints || skillPoints.Value == 0 {
		if ctx.Data.OpenMenus.SkillTree {
			step.CloseAllMenus()
		}
		return nil
	}

	remainingPoints := skillPoints.Value
	skillsToAssign := char.SkillPoints()

	for _, skillID := range skillsToAssign {
		if remainingPoints <= 0 {
			break
		}

		characterPoints, found := ctx.Data.PlayerUnit.Skills[skillID]
		currentLevel := 0
		if found {
			currentLevel = int(characterPoints.Level)
		}

		// Try to spend a point in this skill
		if spendSkillPoint(skillID) {
			remainingPoints--
			ctx.Logger.Debug(fmt.Sprintf("Increased skill %v to level %d (%d total points remaining)",
				skill.SkillNames[skillID], currentLevel+1, remainingPoints))
		} else {
			ctx.Logger.Error(fmt.Sprintf("Failed to spend point in skill %v", skill.SkillNames[skillID]))
		}
	}

	return step.CloseAllMenus()
}

func spendStatPoint(statID stat.ID) bool {
	ctx := context.Get()
	beforePoints, _ := ctx.Data.PlayerUnit.FindStat(stat.StatPoints, 0)

	if !ctx.Data.OpenMenus.Character {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.CharacterScreen)
		utils.Sleep(100)
	}

	statBtnPosition := uiStatButtonPosition[statID]
	if ctx.Data.LegacyGraphics {
		statBtnPosition = uiStatButtonPositionLegacy[statID]
	}

	ctx.HID.Click(game.LeftButton, statBtnPosition.X, statBtnPosition.Y)
	utils.Sleep(300)

	afterPoints, _ := ctx.Data.PlayerUnit.FindStat(stat.StatPoints, 0)
	return beforePoints.Value-afterPoints.Value == 1
}

func spendSkillPoint(skillID skill.ID) bool {
	ctx := context.Get()
	beforePoints, _ := ctx.Data.PlayerUnit.FindStat(stat.SkillPoints, 0)

	if !ctx.Data.OpenMenus.SkillTree {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SkillTree)
		utils.Sleep(100)
	}

	skillDesc, found := skill.Desc[skillID]
	if !found {
		ctx.Logger.Error(fmt.Sprintf("skill not found for character: %v", skillID))
		return false
	}

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
	utils.Sleep(300)

	afterPoints, _ := ctx.Data.PlayerUnit.FindStat(stat.SkillPoints, 0)
	return beforePoints.Value-afterPoints.Value == 1
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

	// Bind right skills
	mainSkill, skillsToBind := char.SkillsToBind()
	if _, found := ctx.Data.Inventory.Find(item.TomeOfTownPortal, item.LocationInventory); found {
		skillsToBind = append(skillsToBind, skill.TomeOfTownPortal)
	}

	notBoundSkills := make([]skill.ID, 0)
	for _, sk := range skillsToBind {
		if _, found := ctx.Data.KeyBindings.KeyBindingForSkill(sk); !found && (sk == skill.TomeOfTownPortal || ctx.Data.PlayerUnit.Skills[sk].Level > 0) {
			notBoundSkills = append(notBoundSkills, sk)
			slices.Sort(notBoundSkills)
			notBoundSkills = slices.Compact(notBoundSkills) // In case we have duplicates (tp tome)
		}
	}

	clvl, _ := ctx.Data.PlayerUnit.FindStat(stat.Level, 0)
	//Hacky way to find if we're lvling a sorc at clvl 1
	str, _ := ctx.Data.PlayerUnit.FindStat(stat.Strength, 0)

	if len(notBoundSkills) > 0 || (clvl.Value == 1 && str.Value == 10) {
		ctx.Logger.Debug("Unbound skills found, trying to bind")
		if ctx.GameReader.LegacyGraphics() {
			ctx.HID.Click(game.LeftButton, ui.SecondarySkillButtonXClassic, ui.SecondarySkillButtonYClassic)
		} else {
			ctx.HID.Click(game.LeftButton, ui.SecondarySkillButtonX, ui.SecondarySkillButtonY)
		}

		availableKB := getAvailableSkillKB()
		ctx.Logger.Debug(fmt.Sprintf("Available KB: %v", availableKB))
		if len(notBoundSkills) > 0 {
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
		} else {
			if _, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.FireBolt); !found {
				ctx.Logger.Debug("Lvl 1 sorc found - forcing fire bolt bind")
				if ctx.GameReader.LegacyGraphics() {
					ctx.HID.MovePointer(1000, 530)
				} else {
					ctx.HID.MovePointer(685, 545)
				}
				utils.Sleep(100)
				ctx.HID.PressKeyBinding(availableKB[0])
				utils.Sleep(300)
			}
		}

	}

	//Set left (main) skill
	if mainSkill > 0 { // If it's anything other than attack
		if ctx.GameReader.LegacyGraphics() {
			ctx.HID.Click(game.LeftButton, ui.MainSkillButtonXClassic, ui.MainSkillButtonYClassic)
		} else {
			ctx.HID.Click(game.LeftButton, ui.MainSkillButtonX, ui.MainSkillButtonY)
		}

		skillPosition, found := calculateSkillPositionInUI(true, mainSkill)
		if found {
			ctx.HID.Click(game.LeftButton, skillPosition.X, skillPosition.Y)
			utils.Sleep(300)
		}
	}
	return nil
}

func calculateSkillPositionInUI(mainSkill bool, skillID skill.ID) (data.Position, bool) {
	ctx := context.Get()

	if _, found := ctx.Data.PlayerUnit.Skills[skillID]; !found {
		return data.Position{}, false
	}

	targetSkill := skill.Skills[skillID]
	descs := make(map[skill.ID]skill.Skill)
	totalRows := make([]int, 0)
	pageSkills := make(map[int][]skill.ID)
	row := 0
	column := 0

	for skID := range ctx.Data.PlayerUnit.Skills {
		sk := skill.Skills[skID]
		// Skip skills that can not be bind
		if sk.Desc().ListRow < 0 {
			continue
		}

		// Skip skills that can not be bind to current mouse button
		if (mainSkill && !sk.LeftSkill) || (!mainSkill && !sk.RightSkill) {
			continue
		}

		descs[skID] = sk
		if sk.Desc().Page == targetSkill.Desc().Page {
			pageSkills[sk.Desc().Page] = append(pageSkills[sk.Desc().Page], skID)
		}
		totalRows = append(totalRows, sk.Desc().ListRow)

	}

	slices.Sort(totalRows)
	totalRows = slices.Compact(totalRows)

	// If we don't have any skill of a specific tree, the entire row gets one line down
	for i, currentRow := range totalRows {
		if currentRow == targetSkill.Desc().ListRow {
			row = i
			break
		}
	}

	skillsInPage := pageSkills[targetSkill.Desc().Page]
	slices.Sort(skillsInPage)
	for i, skills := range skillsInPage {
		if skills == targetSkill.ID {
			column = i
			break
		}
	}

	if ctx.GameReader.LegacyGraphics() {
		skillOffsetX := ui.MainSkillListFirstSkillXClassic + (ui.SkillListSkillOffsetClassic * column)
		if !mainSkill {
			skillOffsetX = ui.SecondarySkillListFirstSkillXClassic - (ui.SkillListSkillOffsetClassic * column)
		}

		return data.Position{
			X: skillOffsetX,
			Y: ui.SkillListFirstSkillYClassic - ui.SkillListSkillOffsetClassic*row,
		}, true
	} else {
		skillOffsetX := ui.MainSkillListFirstSkillX - (ui.SkillListSkillOffset * (len(skillsInPage) - (column + 1)))
		if !mainSkill {
			skillOffsetX = ui.SecondarySkillListFirstSkillX + (ui.SkillListSkillOffset * (len(skillsInPage) - (column + 1))) // Order is reversed in resurrected gfx
		}

		return data.Position{
			X: skillOffsetX,
			Y: ui.SkillListFirstSkillY - ui.SkillListSkillOffset*row,
		}, true
	}
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
			if ctx.Data.LegacyGraphics {
				ctx.HID.Click(game.LeftButton, ui.FirstMercFromContractorListXClassic, ui.FirstMercFromContractorListYClassic)
				utils.Sleep(500)
				ctx.HID.Click(game.LeftButton, ui.FirstMercFromContractorListXClassic, ui.FirstMercFromContractorListYClassic)
			} else {
				ctx.HID.Click(game.LeftButton, ui.FirstMercFromContractorListX, ui.FirstMercFromContractorListY)
				utils.Sleep(500)
				ctx.HID.Click(game.LeftButton, ui.FirstMercFromContractorListX, ui.FirstMercFromContractorListY)
			}
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
