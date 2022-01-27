package helper

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

func ExitGame(actionChan chan<- action.Action, eventChan chan<- event.Event) {
	a := action.NewAction(
		action.PriorityHigh,
		action.NewKeyPress("esc", time.Millisecond*200),
		action.NewMouseDisplacement(time.Millisecond*50, 640, 328),
		action.NewMouseClick(time.Millisecond*120, hid.LeftButton),
	)
	actionChan <- a
	eventChan <- event.ExitedGame
}

func NewGame(actionChan chan<- action.Action, difficulty string) {
	difficultyPosition := map[string]struct {
		X, Y int
	}{
		"normal":    {X: 640, Y: 311},
		"nightmare": {X: 640, Y: 355},
		"hell":      {X: 640, Y: 403},
	}

	createX := difficultyPosition[difficulty].X
	createY := difficultyPosition[difficulty].Y
	a := action.NewAction(
		action.PriorityNormal,
		action.NewMouseDisplacement(time.Millisecond*50, 640, 672),
		action.NewMouseClick(time.Millisecond*350, hid.LeftButton),
		action.NewMouseDisplacement(time.Millisecond*87, createX, createY),
		action.NewMouseClick(time.Millisecond*65, hid.LeftButton),
	)
	actionChan <- a
}
