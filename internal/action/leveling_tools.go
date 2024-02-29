package action

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"log/slog"
	"slices"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
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

var previousTotalSkillNumber = 0

func (b *Builder) EnsureStatPoints() *StepChainAction {
	return NewStepChain(func(d data.Data) []step.Step {
		char, isLevelingChar := b.ch.(LevelingCharacter)
		_, unusedStatPoints := d.PlayerUnit.Stats[stat.StatPoints]
		if !isLevelingChar || !unusedStatPoints {
			if d.OpenMenus.Character {
				return []step.Step{
					step.SyncStep(func(_ data.Data) error {
						b.hid.PressKey("esc")
						return nil
					}),
				}
			}

			return nil
		}

		for st, targetPoints := range char.StatPoints(d) {
			currentPoints, found := d.PlayerUnit.Stats[st]
			if !found || currentPoints >= targetPoints {
				continue
			}

			if !d.OpenMenus.Character {
				return []step.Step{
					step.SyncStep(func(_ data.Data) error {
						b.hid.PressKey(config.Config.Bindings.OpenCharacterScreen)
						return nil
					}),
				}
			}

			statBtnPosition := uiStatButtonPosition[st]
			return []step.Step{
				step.SyncStep(func(_ data.Data) error {
					helper.Sleep(100)
					b.hid.Click(game.LeftButton, statBtnPosition.X, statBtnPosition.Y)
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
	return NewStepChain(func(d data.Data) []step.Step {
		char, isLevelingChar := b.ch.(LevelingCharacter)
		availablePoints, unusedSkillPoints := d.PlayerUnit.Stats[stat.SkillPoints]
		if !isLevelingChar || !unusedSkillPoints || assignAttempts >= availablePoints {
			if d.OpenMenus.SkillTree {
				return []step.Step{
					step.SyncStep(func(_ data.Data) error {
						b.hid.PressKey("esc")
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
					b.logger.Error("skill not found for character", slog.Any("skill", sk))
					return nil
				}

				if !d.OpenMenus.SkillTree {
					return []step.Step{
						step.SyncStep(func(_ data.Data) error {
							b.hid.PressKey(config.Config.Bindings.OpenSkillTree)
							return nil
						}),
					}
				}

				return []step.Step{
					step.SyncStep(func(_ data.Data) error {
						assignAttempts++
						helper.Sleep(100)
						b.hid.Click(game.LeftButton, uiSkillPagePosition[skillDesc.Page-1].X, uiSkillPagePosition[skillDesc.Page-1].Y)
						helper.Sleep(200)
						b.hid.Click(game.LeftButton, uiSkillColumnPosition[skillDesc.Column-1], uiSkillRowPosition[skillDesc.Row-1])
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
	return NewStepChain(func(d data.Data) []step.Step {
		if _, isLevelingChar := b.ch.(LevelingCharacter); !isLevelingChar {
			return nil
		}

		return []step.Step{
			step.SyncStep(func(_ data.Data) error {
				b.hid.PressKey(config.Config.Bindings.OpenQuestLog)
				return nil
			}),
			step.Wait(time.Second),
			step.SyncStep(func(_ data.Data) error {
				b.hid.PressKey(config.Config.Bindings.OpenQuestLog)
				return nil
			}),
		}
	})
}

func (b *Builder) EnsureSkillBindings() *StepChainAction {
	return NewStepChain(func(d data.Data) []step.Step {
		if _, isLevelingChar := b.ch.(LevelingCharacter); !isLevelingChar {
			return nil
		}
		char, _ := b.ch.(LevelingCharacter)
		skillBindings := char.GetKeyBindings(d)
		skillBindings[skill.TomeOfTownPortal] = config.Config.Bindings.TP

		if len(skillBindings) > 0 && len(d.PlayerUnit.Skills) != previousTotalSkillNumber {
			return []step.Step{
				// Right click skill bindings
				step.SyncStep(func(d data.Data) error {
					b.hid.Click(game.LeftButton, ui.SecondarySkillButtonX, ui.SecondarySkillButtonY)
					helper.Sleep(300)
					b.hid.MovePointer(10, 10)
					helper.Sleep(300)

					for sk, binding := range skillBindings {
						if binding == "" {
							continue
						}

						skillPosition, found := b.calculateSkillPositionInUI(d, false, sk)
						if !found {
							continue
						}

						b.hid.MovePointer(skillPosition.X, skillPosition.Y)
						helper.Sleep(100)
						b.hid.PressKey(binding)
						helper.Sleep(300)
					}

					previousTotalSkillNumber = len(d.PlayerUnit.Skills)
					return nil
				}),

				// Set main left click skill
				step.SyncStep(func(_ data.Data) error {
					for sk, binding := range skillBindings {
						if binding != "" {
							continue
						}
						b.hid.Click(game.LeftButton, ui.MainSkillButtonX, ui.MainSkillButtonY)
						helper.Sleep(300)
						b.hid.MovePointer(10, 10)
						helper.Sleep(300)

						skillPosition, found := b.calculateSkillPositionInUI(d, false, sk)
						if !found {
							return nil
						}
						helper.Sleep(100)
						b.hid.Click(game.LeftButton, skillPosition.X, skillPosition.Y)
					}
					return nil
				}),
			}
		}

		return nil
	}, RepeatUntilNoSteps())
}

func (b *Builder) calculateSkillPositionInUI(d data.Data, mainSkill bool, skillID skill.ID) (data.Position, bool) {
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

		if skID != targetSkill.ID && sk.Desc().ListRow == targetSkill.Desc().ListRow && targetSkill.ID < skID {
			column++
		}

		totalRows = append(totalRows, sk.Desc().ListRow)
		if row == targetSkill.Desc().ListRow {
			continue
		}

		row++
	}

	slices.Sort(totalRows)
	totalRows = slices.Compact(totalRows)

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

	return data.Position{
		X: ui.SkillListFirstSkillX + ui.SkillListSkillOffset*column,
		Y: ui.SkillListFirstSkillY - ui.SkillListSkillOffset*row,
	}, true
}

func (b *Builder) HireMerc() *Chain {
	return NewChain(func(d data.Data) (actions []Action) {
		_, isLevelingChar := b.ch.(LevelingCharacter)
		if isLevelingChar && config.Config.Character.UseMerc {
			// Hire the merc if we don't have one, we have enough gold, and we are in act 2. We assume that ReviveMerc was called before this.
			if config.Config.Game.Difficulty == difficulty.Normal && d.MercHPPercent() <= 0 && d.PlayerUnit.TotalGold() > 30000 && d.PlayerUnit.Area == area.LutGholein {
				b.logger.Info("Hiring merc...")
				// TODO: Hire Holy Freeze merc if available, if not, hire Defiance merc.
				actions = append(actions,
					b.InteractNPC(
						town.GetTownByArea(d.PlayerUnit.Area).MercContractorNPC(),
						step.KeySequence("home", "down", "enter"),
						step.Wait(time.Second*2),
						step.SyncStep(func(d data.Data) error {
							b.hid.Click(game.LeftButton, ui.FirstMercFromContractorListX, ui.FirstMercFromContractorListY)
							helper.Sleep(300)
							b.hid.Click(game.LeftButton, ui.FirstMercFromContractorListX, ui.FirstMercFromContractorListY)

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
	return NewChain(func(d data.Data) (actions []Action) {
		ch, isLevelingChar := b.ch.(LevelingCharacter)
		if isLevelingChar && ch.ShouldResetSkills(d) {
			currentArea := d.PlayerUnit.Area
			if d.PlayerUnit.Area != area.RogueEncampment {
				actions = append(actions, b.WayPoint(area.RogueEncampment))
			}
			actions = append(actions,
				b.InteractNPC(npc.Akara,
					step.KeySequence("home", "down", "down", "enter"),
					step.Wait(time.Second),
					step.KeySequence("home", "enter"),
				),
			)
			if d.PlayerUnit.Area != area.RogueEncampment {
				actions = append(actions, b.WayPoint(currentArea))
			}
		}

		return
	})
}
