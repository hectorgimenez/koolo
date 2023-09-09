package action

import (
	"fmt"
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
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/ui"
	"go.uber.org/zap"
)

var uiStatButtonPosition = map[stat.ID]data.Position{
	stat.Strength:  {X: 240, Y: 210},
	stat.Dexterity: {X: 240, Y: 290},
	stat.Vitality:  {X: 240, Y: 380},
	stat.Energy:    {X: 240, Y: 430},
}

var uiSkillTabPosition = []data.Position{
	{X: 910, Y: 140},
	{X: 1010, Y: 140},
	{X: 1100, Y: 140},
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
						hid.PressKey("esc")
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
						hid.PressKey(config.Config.Bindings.OpenCharacterScreen)
						return nil
					}),
				}
			}

			statBtnPosition := uiStatButtonPosition[st]
			return []step.Step{
				step.SyncStep(func(_ data.Data) error {
					helper.Sleep(100)
					hid.MovePointer(statBtnPosition.X, statBtnPosition.Y)
					hid.Click(hid.LeftButton)
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
						hid.PressKey("esc")
						return nil
					}),
				}
			}

			return nil
		}

		skillTree := char.GetSkillTree()
		assignedPoints := make(map[skill.Skill]int)
		for _, sk := range char.SkillPoints(d) {
			currentPoints, found := assignedPoints[sk]
			if !found {
				currentPoints = 0
			}

			assignedPoints[sk] = currentPoints + 1

			characterPoints, found := d.PlayerUnit.Skills[sk]
			if !found || characterPoints < assignedPoints[sk] {
				position, skFound := skillTree[sk]
				if !skFound {
					b.logger.Error("skill not found for character", zap.Any("skill", sk))
					return nil
				}

				if !d.OpenMenus.SkillTree {
					return []step.Step{
						step.SyncStep(func(_ data.Data) error {
							hid.PressKey(config.Config.Bindings.OpenSkillTree)
							return nil
						}),
					}
				}

				return []step.Step{
					step.SyncStep(func(_ data.Data) error {
						assignAttempts++
						helper.Sleep(100)
						hid.MovePointer(uiSkillTabPosition[position.Tab].X, uiSkillTabPosition[position.Tab].Y)
						hid.Click(hid.LeftButton)
						helper.Sleep(200)
						hid.MovePointer(uiSkillColumnPosition[position.Column], uiSkillRowPosition[position.Row])
						hid.Click(hid.LeftButton)
						helper.Sleep(500)
						return nil
					}),
				}
			}
		}

		return nil
	}, RepeatUntilNoSteps())
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
				step.SyncStep(func(_ data.Data) error {
					hid.MovePointer(ui.SecondarySkillButtonX, ui.SecondarySkillButtonY)
					hid.Click(hid.LeftButton)
					helper.Sleep(300)
					hid.MovePointer(10, 10)
					helper.Sleep(300)

					sc := helper.Screenshot()
					for sk, binding := range skillBindings {
						if binding == "" {
							continue
						}
						tm := b.tf.Find(fmt.Sprintf("skills_%d", sk), sc)
						if !tm.Found {
							continue
						}
						hid.MovePointer(tm.PositionX+10, tm.PositionY+10)
						helper.Sleep(100)
						hid.PressKey(binding)
						helper.Sleep(300)
					}

					previousTotalSkillNumber = len(d.PlayerUnit.Skills)
					return nil
				}),
				step.SyncStep(func(_ data.Data) error {
					for sk, binding := range skillBindings {
						if binding != "" {
							continue
						}
						hid.MovePointer(ui.MainSkillButtonX, ui.MainSkillButtonY)
						hid.Click(hid.LeftButton)
						helper.Sleep(300)
						hid.MovePointer(10, 10)
						helper.Sleep(300)

						sc := helper.Screenshot()
						tm := b.tf.Find(fmt.Sprintf("skills_%d", sk), sc)
						if !tm.Found {
							continue
						}
						hid.MovePointer(tm.PositionX+10, tm.PositionY+10)
						helper.Sleep(100)
						hid.Click(hid.LeftButton)
					}
					return nil
				}),
			}
		}

		return nil
	}, RepeatUntilNoSteps())
}

func (b *Builder) GetCompletedQuests(act int) []bool {
	hid.PressKey(config.Config.Bindings.OpenQuestLog)
	hid.MovePointer(ui.QuestFirstTabX+(act-1)*ui.QuestTabXInterval, ui.QuestFirstTabY)
	helper.Sleep(200)
	hid.Click(hid.LeftButton)
	helper.Sleep(3000)

	quests := make([]bool, 6)
	if act == 4 {
		quests = make([]bool, 3)
	}

	sc := helper.Screenshot()
	for i := 0; i < len(quests); i++ {
		tm := b.tf.Find(fmt.Sprintf("quests_a%d_%d", act, i+1), sc)
		quests[i] = tm.Found
	}
	hid.PressKey("esc")

	return quests
}

func (b *Builder) HireMerc() *Chain {
	return NewChain(func(d data.Data) (actions []Action) {
		_, isLevelingChar := b.ch.(LevelingCharacter)
		if isLevelingChar && config.Config.Character.UseMerc {
			// Hire the merc if we don't have one, we have enough gold, and we are in act 2. We assume that ReviveMerc was called before this.
			if config.Config.Game.Difficulty == difficulty.Normal && d.MercHPPercent() <= 0 && d.PlayerUnit.TotalGold() > 30000 && d.PlayerUnit.Area == area.LutGholein {
				b.logger.Info("Hiring merc...")
				actions = append(actions,
					b.InteractNPC(town.GetTownByArea(d.PlayerUnit.Area).MercContractorNPC(), step.KeySequence("home", "down", "enter")),
					NewStepChain(func(d data.Data) []step.Step {
						sc := helper.Screenshot()
						tm := b.tf.Find(fmt.Sprintf("skills_merc_%d", skill.Defiance), sc)
						if !tm.Found {
							return nil
						}

						return []step.Step{
							step.SyncStep(func(d data.Data) error {
								hid.MovePointer(tm.PositionX-100, tm.PositionY)
								hid.Click(hid.LeftButton)
								hid.Click(hid.LeftButton)

								return nil
							}),
						}
					}),
				)
			}

			// We will change the Defiance merc to the holy freeze, much better to avoid being hit.
			if config.Config.Game.Difficulty == difficulty.Nightmare && d.MercHPPercent() > 0 && d.PlayerUnit.TotalGold() > 50000 && d.PlayerUnit.Area == area.KurastDocks && d.PlayerUnit.Skills[skill.Defiance] > 0 {
				b.logger.Info("Changing Defiance merc by Holy Freeze...")
				// Remove merc items
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
