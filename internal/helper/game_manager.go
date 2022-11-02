package helper

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game/difficulty"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/memory"
)

func ExitGame(gr *memory.GameReader) error {
	hid.PressKey("esc")
	Sleep(50)
	hid.PressKey("up")
	Sleep(20)
	hid.PressKey("up")
	Sleep(20)
	hid.PressKey("down")
	Sleep(20)
	hid.PressKey("enter")

	for i := 0; i < 30; i++ {
		if !gr.InGame() {
			return nil
		}
		Sleep(1000)
	}

	return errors.New("error exiting game! Timeout")
}

// TODO: Make this coords dynamic
func NewGame(gr *memory.GameReader) error {
	difficultyPosition := map[difficulty.Difficulty]struct {
		X, Y int
	}{
		difficulty.Normal:    {X: 640, Y: 311},
		difficulty.Nightmare: {X: 640, Y: 355},
		difficulty.Hell:      {X: 640, Y: 403},
	}

	createX := difficultyPosition[config.Config.Game.Difficulty].X
	createY := difficultyPosition[config.Config.Game.Difficulty].Y
	hid.MovePointer(600, 650)
	Sleep(250)
	hid.Click(hid.LeftButton)
	Sleep(250)
	hid.MovePointer(createX, createY)
	hid.Click(hid.LeftButton)

	for i := 0; i < 30; i++ {
		if gr.InGame() {
			return nil
		}
		Sleep(1000)
	}

	return errors.New("error creating game! Timeout")
}
