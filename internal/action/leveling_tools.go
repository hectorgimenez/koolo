package action

import (
	"fmt"
	"log/slog"
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

var baseStats = map[data.Class]map[stat.ID]int{
	data.Amazon:      {stat.Strength: 20, stat.Dexterity: 25, stat.Vitality: 20, stat.Energy: 15},
	data.Assassin:    {stat.Strength: 20, stat.Dexterity: 20, stat.Vitality: 20, stat.Energy: 25},
	data.Barbarian:   {stat.Strength: 30, stat.Dexterity: 20, stat.Vitality: 25, stat.Energy: 10},
	data.Druid:       {stat.Strength: 15, stat.Dexterity: 20, stat.Vitality: 25, stat.Energy: 20},
	data.Necromancer: {stat.Strength: 15, stat.Dexterity: 25, stat.Vitality: 15, stat.Energy: 25},
	data.Paladin:     {stat.Strength: 25, stat.Dexterity: 20, stat.Vitality: 25, stat.Energy: 15},
	data.Sorceress:   {stat.Strength: 10, stat.Dexterity: 25, stat.Vitality: 10, stat.Energy: 35},
}

// #region stat management
type assignedStatPoint struct {
	required int
	assigned int
}

func EnsureStatPoints() error {
	ctx := context.Get()

	char, isLevelingChar := ctx.Char.(context.LevelingCharacter)
	if !isLevelingChar {
		return nil
	}

	unusedStatPoints, hasUnusedStatPoints := ctx.Data.PlayerUnit.FindStat(stat.StatPoints, 0)

	if !hasUnusedStatPoints {
		return nil
	}

	ctx.Logger.Debug("Assigning stat points")

	availableStatPoints := unusedStatPoints.Value
	assignedStatPoints := make(map[stat.ID]*assignedStatPoint)
	for st, totalToAssign := range char.StatPoints() {

		currentStatAssignment, found := assignedStatPoints[st]
		if !found {
			currentStat, _ := ctx.Data.PlayerUnit.BaseStats.FindStat(st, 0)
			currentStatAssignment = &assignedStatPoint{
				required: 0,
				assigned: currentStat.Value - baseStats[ctx.Data.PlayerUnit.Class][st],
			}
			assignedStatPoints[st] = currentStatAssignment
		}

		currentStatAssignment.required += totalToAssign

		for currentStatAssignment.assigned < currentStatAssignment.required && availableStatPoints > 0 {
			if err := clickStatPoint(ctx, st); err != nil {
				ctx.Logger.Error(err.Error(), slog.Any("stat", st))
				continue
			}

			ctx.Logger.Debug("Assigning stat point", slog.Any("stat", stat.StringStats[st]))

			currentStatAssignment.assigned++
			availableStatPoints--
		}
	}

	return nil
}

func clickStatPoint(ctx *context.Status, st stat.ID) error {
	if !ctx.Data.OpenMenus.Character {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.CharacterScreen)
		utils.Sleep(500)
	}

	var statBtnPosition data.Position
	if ctx.Data.LegacyGraphics {
		statBtnPosition = uiStatButtonPositionLegacy[st]
	} else {
		statBtnPosition = uiStatButtonPosition[st]
	}

	ctx.HID.Click(game.LeftButton, statBtnPosition.X, statBtnPosition.Y)
	utils.Sleep(300)

	return nil
}

// #endregion

// #region skill management
type assignedSkill struct {
	skillID        skill.ID
	assignedPoints int
	requiredPoints int
}

func EnsureSkillPoints() error {
	ctx := context.Get()

	char, isLevelingChar := ctx.Char.(context.LevelingCharacter)
	if !isLevelingChar {
		return nil
	}

	availablePointsData, unusedSkillPoints := ctx.Data.PlayerUnit.FindStat(stat.SkillPoints, 0)
	availablePoints := availablePointsData.Value

	if !unusedSkillPoints {
		return nil
	}

	assignedPoints := make(map[skill.ID]*assignedSkill)
	for _, sk := range char.SkillPoints() {
		if availablePoints <= 0 {
			ctx.Logger.Debug("No more skill points available")
			break
		}

		currentPoints, found := assignedPoints[sk]
		if !found {
			currentPoints = &assignedSkill{
				skillID:        sk,
				assignedPoints: int(ctx.Data.PlayerUnit.Skills[sk].Level),
				requiredPoints: 0,
			}

			assignedPoints[sk] = currentPoints
		}

		currentPoints.requiredPoints++

		// If we have already assigned all points for this skill, skip it
		if currentPoints.assignedPoints >= currentPoints.requiredPoints {
			continue
		}

		if err := clickSkill(ctx, sk); err != nil {
			ctx.Logger.Error(err.Error(), slog.Any("skill", sk))
			continue
		}

		ctx.Logger.Debug("Assigning skill point", slog.Any("skill", skill.SkillNames[sk]))

		currentPoints.assignedPoints++
		availablePoints--
	}

	return nil
}

func clickSkill(ctx *context.Status, skillID skill.ID) error {
	skillDesc, found := skill.Desc[skillID]
	if !found {
		return fmt.Errorf("skill not found")
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

	return nil
}

// #endregion

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
