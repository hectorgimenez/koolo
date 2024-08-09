package run

import (
	"slices"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Council struct {
	baseRun
}

func (s Council) Name() string {
	return string(config.CouncilRun)
}

func (s Council) BuildActions() (actions []action.Action) {
	monsterFilter := data.MonsterAnyFilter()

	actions = append(actions,
		s.builder.WayPoint(area.Travincal),
		s.builder.Buff(),
	)

	chainActions := action.NewChain(func(d game.Data) []action.Action {
		var actions []action.Action

		playerPosX := d.PlayerUnit.Position.X
		playerPosY := d.PlayerUnit.Position.Y

		// Setting original position to start from
		originalPosition := data.Position{X: playerPosX, Y: playerPosY}

		// Move outside of WP building
		TraviMoveCoords1 := data.Position{X: (originalPosition.X + 18), Y: originalPosition.Y}
		actions = append(actions, s.builder.MoveToCoords(TraviMoveCoords1))
		originalPosition = TraviMoveCoords1

		// Move up towards starting point
		TraviMoveCoords2 := data.Position{X: originalPosition.X, Y: (originalPosition.Y - 27)}
		actions = append(actions, s.builder.MoveToCoords(TraviMoveCoords2))
		originalPosition = TraviMoveCoords2

		// Move to clear position 1, ignoring water snakes
		TraviClearCoords1 := data.Position{X: (originalPosition.X + 80), Y: originalPosition.Y}
		actions = append(actions,
			s.builder.MoveToCoords(TraviClearCoords1),
			s.builder.ClearAreaAroundPlayer(10, func(m data.Monsters) []data.Monster {
				var monsters []data.Monster
				monsters = monsterFilter(m)
				monsters = skipTraviMonstersFilter(monsters)
				return monsters
			}),
			s.builder.ItemPickup(false, 20),
		)
		originalPosition = TraviClearCoords1

		// Move to clear position 2, ignoring water snakes
		TraviClearCoords2 := data.Position{X: originalPosition.X, Y: (originalPosition.Y - 24)}
		actions = append(actions,
			s.builder.MoveToCoords(TraviClearCoords2),
			s.builder.ClearAreaAroundPlayer(20, func(m data.Monsters) []data.Monster {
				var monsters []data.Monster
				monsters = monsterFilter(m)
				monsters = skipTraviMonstersFilter(monsters)
				return monsters
			}),
			s.builder.ItemPickup(false, 20),
		)
		originalPosition = TraviClearCoords2

		// Move to clear position 3, killing council members
		TraviClearCoords3 := data.Position{X: originalPosition.X, Y: (originalPosition.Y - 23)}
		actions = append(actions,
			s.builder.MoveToCoords(TraviClearCoords3),
			s.builder.ClearAreaAroundPlayer(20, func(m data.Monsters) []data.Monster {
				var monsters []data.Monster
				monsters = monsterFilter(m)
				monsters = targetCouncilMembers(monsters)
				return monsters
			}),
			s.builder.ItemPickup(false, 20),
		)
		originalPosition = TraviClearCoords3

		// Move to clear position 4, killing council members
		TraviClearCoords4 := data.Position{X: originalPosition.X, Y: (originalPosition.Y - 20)}
		actions = append(actions,
			s.builder.MoveToCoords(TraviClearCoords4),
			s.builder.ClearAreaAroundPlayer(20, func(m data.Monsters) []data.Monster {
				var monsters []data.Monster
				monsters = monsterFilter(m)
				monsters = targetCouncilMembers(monsters)
				return monsters
			}),
			s.builder.ItemPickup(false, 20),
		)
		originalPosition = TraviClearCoords4

		return actions
	})

	actions = append(actions, chainActions)

	return actions
}

// Target only councilmembers
func targetCouncilMembers(monsters []data.Monster) []data.Monster {
	var councilMembers []data.Monster
	for _, m := range monsters {
		if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
			councilMembers = append(councilMembers, m)
		}
	}
	return councilMembers
}

// Skip the water snakes else you might get stuck on them
func skipTraviMonstersFilter(monsters data.Monsters) []data.Monster {
	var waterWatcherHeadIds = []npc.ID{
		npc.WaterWatcherHead,
	}
	var filteredMonsters []data.Monster

	for _, m := range monsters {
		if !slices.Contains(waterWatcherHeadIds, m.Name) {
			filteredMonsters = append(filteredMonsters, m)
		}
	}

	return filteredMonsters
}
