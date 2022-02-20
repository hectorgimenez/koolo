package helper

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/hid"
)

func ExitGame() {
	hid.PressKey("esc")
	Sleep(150)
	hid.PressKey("up")
	Sleep(50)
	hid.PressKey("up")
	Sleep(50)
	hid.PressKey("down")
	Sleep(50)
	hid.PressKey("enter")
}

// TODO: Make this coords dynamic
func NewGame() {
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
}
