package step

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type MoveToAreaStep struct {
	pathingStep
	area                  area.Area
	waitingForInteraction bool
}

func MoveToLevel(area area.Area) *MoveToAreaStep {
	return &MoveToAreaStep{
		pathingStep: newPathingStep(),
		area:        area,
	}
}

func (m *MoveToAreaStep) Status(d data.Data) Status {
	if m.status == StatusCompleted {
		return StatusCompleted
	}

	// Give some extra time to render the UI
	if d.PlayerUnit.Area == m.area && time.Since(m.lastRun) > time.Second*1 {
		return m.tryTransitionStatus(StatusCompleted)
	}

	return m.status
}

func (m *MoveToAreaStep) Run(d data.Data) error {
	if m.status == StatusNotStarted && CanTeleport(d) {
		hid.PressKey(config.Config.Bindings.Teleport)
	}
	m.tryTransitionStatus(StatusInProgress)
	if CanTeleport(d) && time.Since(m.lastRun) < config.Config.Runtime.CastDuration {
		return nil
	}

	if (m.waitingForInteraction && time.Since(m.lastRun) < time.Second*1) || d.PlayerUnit.Area == m.area {
		return nil
	}

	m.lastRun = time.Now()
	for _, l := range d.AdjacentLevels {
		if l.Area == m.area {
			distance := pather.DistanceFromMe(d, l.Position)
			if distance > 5 {
				stuck := m.isPlayerStuck(d)
				if m.path == nil || ! m.cachePath(d, m.expandedCG) || stuck {
					if stuck {
						tile := m.path.AstarPather[m.path.Distance()-1].(*pather.Tile)
						m.blacklistedPositions = append(m.blacklistedPositions, [2]int{tile.X, tile.Y})
					}
					expandedCG := pather.ShouldExpandCollisionGrid(d, l.Position)
					path, _, found := pather.GetPath(d, l.Position)
					if !found {
						return errors.New("path could not be calculated, maybe there is an obstacle")
					}
					m.path = path
					m.expandedCG = expandedCG
				}
				pather.MoveThroughPath(m.path, calculateMaxDistance(d), CanTeleport(d))
				return nil
			}

			if l.CanInteract {
				x, y := pather.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, l.Position.X-2, l.Position.Y-2)
				hid.MovePointer(x, y)
				helper.Sleep(100)
				hid.Click(hid.LeftButton)
				m.waitingForInteraction = true
				return nil
			} else {
				hid.PressKey(config.Config.Bindings.ForceMove)
			}
		}
	}

	return fmt.Errorf("area %s not found", m.area)
}

func CanTeleport(d data.Data) bool {
	_, found := d.PlayerUnit.Skills[skill.Teleport]

	return found && config.Config.Bindings.Teleport != "" && !d.PlayerUnit.Area.IsTown()
}

func calculateMaxDistance(d data.Data) int {
	if CanTeleport(d) {
		return 25
	}

	return 15
}
