package action

import (
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

func EnsureStatPoints() error {
	// TODO finish this
	return nil
	//return NewStepChain(func(d game.Data) []step.Step {
	//	char, isLevelingChar := b.ch.(LevelingCharacter)
	//	_, unusedStatPoints := d.PlayerUnit.FindStat(stat.StatPoints, 0)
	//	if !isLevelingChar || !unusedStatPoints {
	//		if d.OpenMenus.Character {
	//			return []step.Step{
	//				step.SyncStep(func(_ game.Data) error {
	//					b.HID.PressKey(win.VK_ESCAPE)
	//					return nil
	//				}),
	//			}
	//		}
	//
	//		return nil
	//	}
	//
	//	for st, targetPoints := range char.StatPoints(d) {
	//		currentPoints, found := d.PlayerUnit.FindStat(st, 0)
	//		if !found || currentPoints.Value >= targetPoints {
	//			continue
	//		}
	//
	//		if !d.OpenMenus.Character {
	//			return []step.Step{
	//				step.SyncStep(func(_ game.Data) error {
	//					b.HID.PressKeyBinding(d.KeyBindings.CharacterScreen)
	//					return nil
	//				}),
	//			}
	//		}
	//
	//		var statBtnPosition data.Position
	//		if d.LegacyGraphics {
	//			statBtnPosition = uiStatButtonPositionLegacy[st]
	//		} else {
	//			statBtnPosition = uiStatButtonPosition[st]
	//		}
	//
	//		return []step.Step{
	//			step.SyncStep(func(_ game.Data) error {
	//				utils.Sleep(100)
	//				b.HID.Click(game.LeftButton, statBtnPosition.X, statBtnPosition.Y)
	//				utils.Sleep(500)
	//				return nil
	//			}),
	//		}
	//	}
	//
	//	return nil
	//}, RepeatUntilNoSteps())
}

func EnsureSkillPoints() error {
	// TODO finish this
	return nil
	//ctx := context.Get()
	//
	//char, isLevelingChar := ctx.Char.(LevelingCharacter)
	//availablePoints, unusedSkillPoints := ctx.Data.PlayerUnit.FindStat(stat.SkillPoints, 0)
	//
	//assignedPoints := make(map[skill.ID]int)
	//for _, sk := range char.SkillPoints() {
	//	currentPoints, found := assignedPoints[sk]
	//	if !found {
	//		currentPoints = 0
	//	}
	//
	//	assignedPoints[sk] = currentPoints + 1
	//
	//	characterPoints, found := ctx.Data.PlayerUnit.Skills[sk]
	//	if !found || int(characterPoints.Level) < assignedPoints[sk] {
	//		skillDesc, skFound := skill.Desc[sk]
	//		if !skFound {
	//			ctx.Logger.Error("skill not found for character", slog.Any("skill", sk))
	//			return nil
	//		}
	//
	//		if !ctx.Data.OpenMenus.SkillTree {
	//			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.SkillTree)
	//		}
	//
	//		utils.Sleep(100)
	//		if ctx.Data.LegacyGraphics {
	//			ctx.HID.Click(game.LeftButton, uiSkillPagePositionLegacy[skillDesc.Page-1].X, uiSkillPagePositionLegacy[skillDesc.Page-1].Y)
	//		} else {
	//			ctx.HID.Click(game.LeftButton, uiSkillPagePosition[skillDesc.Page-1].X, uiSkillPagePosition[skillDesc.Page-1].Y)
	//		}
	//		utils.Sleep(200)
	//		if ctx.Data.LegacyGraphics {
	//			ctx.HID.Click(game.LeftButton, uiSkillColumnPositionLegacy[skillDesc.Column-1], uiSkillRowPositionLegacy[skillDesc.Row-1])
	//		} else {
	//			ctx.HID.Click(game.LeftButton, uiSkillColumnPosition[skillDesc.Column-1], uiSkillRowPosition[skillDesc.Row-1])
	//		}
	//		utils.Sleep(500)
	//		return nil
	//	}
	//}
	//
	//return nil
}

func UpdateQuestLog() error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "UpdateQuestLog"

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
	ctx.ContextDebug.LastStep = "getAvailableSkillKB"

	for _, sb := range ctx.Data.KeyBindings.Skills {
		if sb.SkillID == -1 && (sb.Key1[0] != 0 && sb.Key1[0] != 255) || (sb.Key2[0] != 0 && sb.Key2[0] != 255) {
			availableSkillKB = append(availableSkillKB, sb.KeyBinding)
		}
	}

	return availableSkillKB
}

func EnsureSkillBindings() error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "EnsureSkillBindings"

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
	ctx.ContextDebug.LastAction = "HireMerc"

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
	ctx.ContextDebug.LastAction = "ResetStats"

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
	ctx.ContextDebug.LastAction = "WaitForAllMembersWhenLeveling"

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
