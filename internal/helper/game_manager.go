package helper

import (
	"context"
	"errors"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
)

func ExitGame(ctx context.Context) error {
	hid.PressKey("esc")
	Sleep(150)
	hid.PressKey("up")
	Sleep(50)
	hid.PressKey("up")
	Sleep(50)
	hid.PressKey("down")
	Sleep(50)
	hid.PressKey("enter")

	for i := 0; i < 30; i++ {
		if game.Status(ctx).Area == "" {
			return nil
		}
		Sleep(1000)
	}

	return errors.New("error exiting game! Timeout")
}

// TODO: Make this coords dynamic
func NewGame(ctx context.Context) error {
	difficultyPosition := map[string]struct {
		X, Y int
	}{
		"normal":    {X: 640, Y: 311},
		"nightmare": {X: 640, Y: 355},
		"hell":      {X: 640, Y: 403},
	}

	createX := difficultyPosition[config.Config.Character.Difficulty].X
	createY := difficultyPosition[config.Config.Character.Difficulty].Y
	hid.MovePointer(640, 672)
	Sleep(50)
	hid.Click(hid.LeftButton)
	Sleep(200)
	hid.MovePointer(createX, createY)
	hid.Click(hid.LeftButton)

	for i := 0; i < 30; i++ {
		if game.Status(ctx).Area != "" {
			return nil
		}
		Sleep(1000)
	}

	return errors.New("error creating game! Timeout")
}
