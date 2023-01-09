package helper

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game/difficulty"
	"github.com/hectorgimenez/koolo/internal/helper/tcp"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/memory"
)

// ExitGame tries to close the socket and also exit via game menu, what happens faster.
func ExitGame(gr *memory.GameReader) error {
	_ = tcp.CloseCurrentGameSocket(gr.GetPID())
	exitGameUsingUIMenu()

	for i := 0; i < 30; i++ {
		if !gr.InGame() {
			return nil
		}
		Sleep(1000)
	}

	return errors.New("error exiting game! Timeout")
}

func exitGameUsingUIMenu() {
	hid.PressKey("esc")
	hid.MovePointer(hid.GameAreaSizeX/2, int(float64(hid.GameAreaSizeY)/2.2))
	hid.Click(hid.LeftButton)
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
