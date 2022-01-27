package game

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type GameManager struct {
	cfg        config.Config
	actionChan chan<- action.Action
	eventChan  chan<- event.Event
}

func NewGameManager(cfg config.Config, actionChan chan<- action.Action, eventChan chan<- event.Event) GameManager {
	return GameManager{
		cfg:        cfg,
		actionChan: actionChan,
		eventChan:  eventChan,
	}
}

func (gm GameManager) ExitGame() {
	a := action.NewAction(
		action.PriorityHigh,
		action.NewKeyPress("esc", time.Millisecond*200),
		action.NewMouseDisplacement(time.Millisecond*50, 640, 328),
		action.NewMouseClick(time.Millisecond*120, hid.LeftButton),
	)
	gm.actionChan <- a
	gm.eventChan <- event.ExitedGame
}

func (gm GameManager) NewGame() {
	difficultyPosition := map[string]struct {
		X, Y int
	}{
		"normal":    {X: 640, Y: 311},
		"nightmare": {X: 640, Y: 355},
		"hell":      {X: 640, Y: 403},
	}

	createX := difficultyPosition[gm.cfg.Character.Difficulty].X
	createY := difficultyPosition[gm.cfg.Character.Difficulty].Y
	a := action.NewAction(
		action.PriorityNormal,
		action.NewMouseDisplacement(time.Millisecond*50, 640, 672),
		action.NewMouseClick(time.Millisecond*350, hid.LeftButton),
		action.NewMouseDisplacement(time.Millisecond*87, createX, createY),
		action.NewMouseClick(time.Millisecond*65, hid.LeftButton),
	)
	gm.actionChan <- a
}
