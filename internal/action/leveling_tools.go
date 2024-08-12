package action

import (
	"log/slog"
	"slices"
	"time"

	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/lxn/win"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/ui"
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

func (b *Builder) EnsureStatPoints() *StepChainAction {
	return NewStepChain(func(d game.Data) []step.Step {
		char, isLevelingChar := b.ch.(LevelingCharacter)
		_, unusedStatPoints := d.PlayerUnit.FindStat(stat.StatPoints, 0)
		if !isLevelingChar || !unusedStatPoints {
			if d.OpenMenus.Character {
				return []step.Step{
					step.SyncStep(func(_ game.Data) error {
						b.HID.PressKey(win.VK_ESCAPE)
						return nil
					}),
				}
			}

			return nil
		}

		for st, targetPoints := range char.StatPoints(d) {
			currentPoints, found := d.PlayerUnit.FindStat(st, 0)
			if !found || currentPoints.Value >= targetPoints {
				continue
			}

			if !d.OpenMenus.Character {
				return []step.Step{
					step.SyncStep(func(_ game.Data) error {
						b.HID.PressKeyBinding(d.KeyBindings.CharacterScreen)
						return nil
					}),
				}
			}

			var statBtnPosition data.Position
			if d.LegacyGraphics {
				statBtnPosition = uiStatButtonPositionLegacy[st]
			} else {
				statBtnPosition = uiStatButtonPosition[st]
			}

			return []step.Step{
				step.SyncStep(func(_ game.Data) error {
					helper.Sleep(100)
					b.HID.Click(game.LeftButton, statBtnPosition.X, statBtnPosition.Y)
					helper.Sleep(500)
					return nil
				}),
			}
		}

		return nil
	}, RepeatUntilNoSteps())
}

func (b *Builder) EnsureSkillPoints() *StepChainAction {
	assignAttempts := 0
	return NewStepChain(func(d game.Data) []step.Step {
		char, isLevelingChar := b.ch.(LevelingCharacter)
		availablePoints, unusedSkillPoints := d.PlayerUnit.FindStat(stat.SkillPoints, 0)
		if !isLevelingChar || !unusedSkillPoints || assignAttempts >= availablePoints.Value {
			if d.OpenMenus.SkillTree {
				return []step.Step{
					step.SyncStep(func(_ game.Data) error {
						b.HID.PressKey(win.VK_ESCAPE)
						return nil
					}),
				}
			}

			return nil
		}

		assignedPoints := make(map[skill.ID]int)
		for _, sk := range char.SkillPoints(d) {
			currentPoints, found := assignedPoints[sk]
			if !found {
				currentPoints = 0
			}

			assignedPoints[sk] = currentPoints + 1

			characterPoints, found := d.PlayerUnit.Skills[sk]
			if !found || int(characterPoints.Level) < assignedPoints[sk] {
				skillDesc, skFound := skill.Desc[sk]
				if !skFound {
					b.Logger.Error("skill not found for character", slog.Any("skill", sk))
					return nil
				}

				if !d.OpenMenus.SkillTree {
					return []step.Step{
						step.SyncStep(func(_ game.Data) error {
							b.HID.PressKeyBinding(d.KeyBindings.SkillTree)
							return nil
						}),
					}
				}

				return []step.Step{
					step.SyncStep(func(_ game.Data) error {
						assignAttempts++
						helper.Sleep(100)
						if d.LegacyGraphics {
							b.HID.Click(game.LeftButton, uiSkillPagePositionLegacy[skillDesc.Page-1].X, uiSkillPagePositionLegacy[skillDesc.Page-1].Y)
						} else {
							b.HID.Click(game.LeftButton, uiSkillPagePosition[skillDesc.Page-1].X, uiSkillPagePosition[skillDesc.Page-1].Y)
						}
						helper.Sleep(200)
						if d.LegacyGraphics {
							b.HID.Click(game.LeftButton, uiSkillColumnPositionLegacy[skillDesc.Column-1], uiSkillRowPositionLegacy[skillDesc.Row-1])
						} else {
							b.HID.Click(game.LeftButton, uiSkillColumnPosition[skillDesc.Column-1], uiSkillRowPosition[skillDesc.Row-1])
						}
						helper.Sleep(500)
						return nil
					}),
				}
			}
		}

		return nil
	}, RepeatUntilNoSteps())
}

func (b *Builder) UpdateQuestLog() *StepChainAction {
	return NewStepChain(func(d game.Data) []step.Step {
		if _, isLevelingChar := b.ch.(LevelingCharacter); !isLevelingChar {
			return nil
		}

		return []step.Step{
			step.SyncStep(func(_ game.Data) error {
				b.HID.PressKeyBinding(d.KeyBindings.QuestLog)
				return nil
			}),
			step.Wait(time.Second),
			step.SyncStep(func(_ game.Data) error {
				b.HID.PressKeyBinding(d.KeyBindings.QuestLog)
				return nil
			}),
		}
	})
}
func (b *Builder) getAvailableSkillKB(d game.Data) []data.KeyBinding {
	availableSkillKB := make([]data.KeyBinding, 0)

	for _, sb := range d.KeyBindings.Skills {
		if sb.SkillID == -1 && (sb.Key1[0] != 0 && sb.Key1[0] != 255) || (sb.Key2[0] != 0 && sb.Key2[0] != 255) {
			availableSkillKB = append(availableSkillKB, sb.KeyBinding)
		}
	}

	return availableSkillKB
}

func (b *Builder) EnsureSkillBindings() *StepChainAction {
	return NewStepChain(func(d game.Data) (steps []step.Step) {
		if _, isLevelingChar := b.ch.(LevelingCharacter); !isLevelingChar {
			return nil
		}
		char, _ := b.ch.(LevelingCharacter)
		mainSkill, skillsToBind := char.SkillsToBind(d)
		skillsToBind = append(skillsToBind, skill.TomeOfTownPortal)
		notBoundSkills := make([]skill.ID, 0)
		for _, sk := range skillsToBind {
			if _, found := d.KeyBindings.KeyBindingForSkill(sk); !found && d.PlayerUnit.Skills[sk].Level > 0 {
				notBoundSkills = append(notBoundSkills, sk)
			}
		}

		if len(notBoundSkills) > 0 {
			steps = append(steps, step.SyncStep(func(d game.Data) error {
				b.HID.Click(game.LeftButton, ui.SecondarySkillButtonX, ui.SecondarySkillButtonY)
				helper.Sleep(300)
				b.HID.MovePointer(10, 10)
				helper.Sleep(300)

				availableKB := b.getAvailableSkillKB(d)

				for i, sk := range notBoundSkills {
					skillPosition, found := b.calculateSkillPositionInUI(d, false, sk)
					if !found {
						continue
					}

					b.HID.MovePointer(skillPosition.X, skillPosition.Y)
					helper.Sleep(100)
					b.HID.PressKeyBinding(availableKB[i])
					helper.Sleep(300)
				}

				return nil
			}))
		}

		if d.PlayerUnit.LeftSkill != mainSkill {
			steps = append(steps, step.SyncStep(func(d game.Data) error {
				b.HID.Click(game.LeftButton, ui.MainSkillButtonX, ui.MainSkillButtonY)
				helper.Sleep(300)
				b.HID.MovePointer(10, 10)
				helper.Sleep(300)

				skillPosition, found := b.calculateSkillPositionInUI(d, true, mainSkill)
				if found {
					b.HID.MovePointer(skillPosition.X, skillPosition.Y)
					helper.Sleep(100)
					b.HID.Click(game.LeftButton, skillPosition.X, skillPosition.Y)
					helper.Sleep(300)
				}

				return nil
			}))
		}

		return
	}, RepeatUntilNoSteps())
}

func (b *Builder) calculateSkillPositionInUI(d game.Data, mainSkill bool, skillID skill.ID) (data.Position, bool) {
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

func (b *Builder) HireMerc() *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		_, isLevelingChar := b.ch.(LevelingCharacter)
		if isLevelingChar && d.CharacterCfg.Character.UseMerc {
			// Hire the merc if we don't have one, we have enough gold, and we are in act 2. We assume that ReviveMerc was called before this.
			if d.CharacterCfg.Game.Difficulty == difficulty.Normal && d.MercHPPercent() <= 0 && d.PlayerUnit.TotalPlayerGold() > 30000 && d.PlayerUnit.Area == area.LutGholein {
				b.Logger.Info("Hiring merc...")
				// TODO: Hire Holy Freeze merc if available, if not, hire Defiance merc.
				actions = append(actions,
					b.InteractNPC(
						town.GetTownByArea(d.PlayerUnit.Area).MercContractorNPC(),
						step.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN),
						step.Wait(time.Second*2),
						step.SyncStep(func(d game.Data) error {
							b.HID.Click(game.LeftButton, ui.FirstMercFromContractorListX, ui.FirstMercFromContractorListY)
							helper.Sleep(300)
							b.HID.Click(game.LeftButton, ui.FirstMercFromContractorListX, ui.FirstMercFromContractorListY)

							return nil
						}),
					),
				)
			}
		}

		return
	})
}

func (b *Builder) ResetStats() *Chain {
	return NewChain(func(d game.Data) (actions []Action) {
		ch, isLevelingChar := b.ch.(LevelingCharacter)
		if isLevelingChar && ch.ShouldResetSkills(d) {
			currentArea := d.PlayerUnit.Area
			if d.PlayerUnit.Area != area.RogueEncampment {
				actions = append(actions, b.WayPoint(area.RogueEncampment))
			}
			actions = append(actions,
				b.InteractNPC(npc.Akara,
					step.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_DOWN, win.VK_RETURN),
					step.Wait(time.Second),
					step.KeySequence(win.VK_HOME, win.VK_RETURN),
				),
			)
			if d.PlayerUnit.Area != area.RogueEncampment {
				actions = append(actions, b.WayPoint(currentArea))
			}
		}

		return
	})
}

func (b *Builder) WaitForAllMembersWhenLeveling() *Chain {
	return NewChain(func(d game.Data) []Action {
		_, isLeveling := b.ch.(LevelingCharacter)
		if d.CharacterCfg.Companion.Enabled && d.CharacterCfg.Companion.Leader && !d.PlayerUnit.Area.IsTown() && isLeveling {
			allMembersAreaCloseToMe := true
			for _, member := range d.Roster {
				if member.Name != d.PlayerUnit.Name && pather.DistanceFromMe(d, member.Position) > 20 {
					allMembersAreaCloseToMe = false
				}
			}

			if allMembersAreaCloseToMe {
				return nil
			}

			return []Action{b.ClearAreaAroundPlayer(5, data.MonsterAnyFilter())}
		}

		return nil
	}, RepeatUntilNoSteps())
}
