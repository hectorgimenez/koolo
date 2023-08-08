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
	mouseOverAttempts     int
	pathFindAttempts	  int
}

var last_x int = 0
var last_y int = 0
var last_tp_attmpt time.Time

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

	if !CanTeleport(d) && time.Since(m.lastRun) < helper.RandomDurationMs(400, 600) {
		return nil
	}

	if CanTeleport(d) && time.Since(m.lastRun) < config.Config.Runtime.CastDuration {
		return nil
	}

	if m.mouseOverAttempts > maxInteractions {
		return fmt.Errorf("area %d could not be interacted", m.area)
	}

	if (m.waitingForInteraction && time.Since(m.lastRun) < time.Second*1) || d.PlayerUnit.Area == m.area {
		return nil
	}

	m.lastRun = time.Now()
	for _, l := range d.AdjacentLevels {
		if l.Area == m.area {
			distance := pather.DistanceFromMe(d, l.Position)
			destination := l.Position
			can_tele := CanTeleport(d)
			since_last_tp :=  time.Now().Sub(last_tp_attmpt)
			if distance < 40 && can_tele && since_last_tp.Milliseconds() < 500 {
				// if last position is the same as current position, dont do tele
				if last_x == d.PlayerUnit.Position.X && last_y == d.PlayerUnit.Position.Y {
					fmt.Println(fmt.Sprintf("%s My position has not changed since last tp %d, dont tele", time.Now(), since_last_tp.Milliseconds()))
					return nil
				}
			}

			if distance > 5 {
				stuck := m.isPlayerStuck(d)
				if m.path == nil || !m.cachePath(d) || stuck {
					if stuck {
						tile := m.path.AstarPather[m.path.Distance()-1].(*pather.Tile)
						m.blacklistedPositions = append(m.blacklistedPositions, [2]int{tile.X, tile.Y})
					}
					if m.pathFindAttempts > 0 {
						var x_diff float64 = float64(l.Position.X - d.PlayerUnit.Position.X)
						var y_diff float64 = float64(l.Position.Y - d.PlayerUnit.Position.Y)
						destination.X = int(x_diff * float64(10 - m.pathFindAttempts) / 10.0) + d.PlayerUnit.Position.X
						destination.Y = int(y_diff * float64(10 - m.pathFindAttempts) / 10.0) + d.PlayerUnit.Position.Y
					}
					path, _, found := pather.GetPath(d, destination)
					if !found {
						m.pathFindAttempts++
						return errors.New("path could not be calculated, maybe there is an obstacle")
					}
					m.path = path
				}
				pather.MoveThroughPath(m.path, calculateMaxDistance(d), can_tele)

				// otherwise go ahead and tele, record position
				last_x = d.PlayerUnit.Position.X
				last_y = d.PlayerUnit.Position.Y
				last_tp_attmpt = time.Now()
				
				return nil
			}

			if l.IsEntrance {
				if d.HoverData.UnitType == 5 || d.HoverData.UnitType == 2 && d.HoverData.IsHovered {
					hid.Click(hid.LeftButton)
					m.waitingForInteraction = true
				}

				lx, ly := pather.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, l.Position.X-2, l.Position.Y-2)
				x, y := helper.Spiral(m.mouseOverAttempts)
				hid.MovePointer(lx+x, ly+y)
				m.mouseOverAttempts++
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

	// Duriel's Lair is bugged and teleport doesn't work here
	if d.PlayerUnit.Area == area.DurielsLair {
		return false
	}

	return found && config.Config.Bindings.Teleport != "" && !d.PlayerUnit.Area.IsTown()
}

func calculateMaxDistance(d data.Data) int {
	if CanTeleport(d) {
		return 25
	}

	return helper.RandRng(7, 11)
}
