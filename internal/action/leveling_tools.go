package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
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


func (b Builder) EnsureEmptyHand() *StaticAction {
	return BuildStatic(func(d data.Data) (steps []step.Step) {
		hid.Click(hid.LeftButton)
		return nil
	})
}

func (b Builder) EnsureStatPoints() *DynamicAction {
	return BuildDynamic(func(d data.Data) ([]step.Step, bool) {
		char, isLevelingChar := b.ch.(LevelingCharacter)
		_, unusedStatPoints := d.PlayerUnit.Stats[stat.StatPoints]
		if !isLevelingChar || !unusedStatPoints {
			if d.OpenMenus.Character {
				return []step.Step{
					step.SyncStep(func(_ data.Data) error {
						hid.PressKey("esc")
						return nil
					}),
				}, true
			}

			return nil, false
		}

		for st, targetPoints := range char.StatPoints() {
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
				}, true
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
			}, true
		}

		return nil, false
	}, CanBeSkipped())
}

func (b Builder) EnsureSkillPoints() *DynamicAction {
	return BuildDynamic(func(d data.Data) ([]step.Step, bool) {
		char, isLevelingChar := b.ch.(LevelingCharacter)
		_, unusedSkillPoints := d.PlayerUnit.Stats[stat.SkillPoints]
		if !isLevelingChar || !unusedSkillPoints {
			if d.OpenMenus.SkillTree {
				return []step.Step{
					step.SyncStep(func(_ data.Data) error {
						hid.PressKey("esc")
						return nil
					}),
				}, true
			}

			return nil, false
		}

		assignedPoints := make(map[skill.Skill]int, 0)
		for _, sk := range char.SkillPoints() {
			currentPoints, found := assignedPoints[sk]
			if !found {
				currentPoints = 0
			}

			assignedPoints[sk] = currentPoints + 1

			characterPoints, found := d.PlayerUnit.Skills[sk]
			if !found || characterPoints < assignedPoints[sk] {
				position, skFound := skill.SorceressTree[sk]
				if !skFound {
					b.logger.Error("skill not found for character", zap.Any("skill", sk))
					return nil, false
				}

				if !d.OpenMenus.SkillTree {
					return []step.Step{
						step.SyncStep(func(_ data.Data) error {
							hid.PressKey(config.Config.Bindings.OpenSkillTree)
							return nil
						}),
					}, true
				}

				return []step.Step{
					step.SyncStep(func(_ data.Data) error {
						helper.Sleep(100)
						hid.MovePointer(uiSkillTabPosition[position.Tab].X, uiSkillTabPosition[position.Tab].Y)
						hid.Click(hid.LeftButton)
						helper.Sleep(200)
						hid.MovePointer(uiSkillColumnPosition[position.Column], uiSkillRowPosition[position.Row])
						hid.Click(hid.LeftButton)
						helper.Sleep(500)
						return nil
					}),
				}, true
			}
		}

		return nil, false
	}, CanBeSkipped())
}
