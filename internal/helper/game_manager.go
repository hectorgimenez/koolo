package helper

import (
	"errors"
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/reader"
)

type GameManager struct {
	gr *reader.GameReader
}

func NewGameManager(gr *reader.GameReader) *GameManager {
	return &GameManager{gr: gr}
}

// ExitGame tries to close the socket and also exit via game menu, what happens faster.
func (gm *GameManager) ExitGame() error {
	//_ = tcp.CloseCurrentGameSocket(gm.gr.GetPID())
	exitGameUsingUIMenu()

	for i := 0; i < 30; i++ {
		if !gm.gr.InGame() {
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

func (gm *GameManager) NewGame() error {
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
		if gm.gr.InGame() {
			return nil
		}
		Sleep(1000)
	}

	return errors.New("error creating game! Timeout")
}

func (gm *GameManager) CreateOnlineGame(gameCounter int) (string, error) {
	// Enter bnet lobby
	hid.MovePointer(744, 650)
	hid.Click(hid.LeftButton)
	Sleep(1200)

	// Click "Create game" tab
	hid.MovePointer(845, 54)
	hid.Click(hid.LeftButton)
	Sleep(200)

	// Click the game name textbox, delete text and type new game name
	hid.MovePointer(925, 116)
	hid.Click(hid.LeftButton)
	hid.PressKeyCombination("lctrl", "a")
	gameName := config.Config.Companion.GameNameTemplate + fmt.Sprintf("%d", gameCounter)
	for _, ch := range gameName {
		hid.PressKey(fmt.Sprintf("%c", ch))
	}

	// Same for password
	hid.MovePointer(925, 161)
	hid.Click(hid.LeftButton)
	Sleep(200)
	hid.PressKeyCombination("lctrl", "a")
	hid.PressKey("x")
	hid.PressKey("enter")

	for i := 0; i < 30; i++ {
		if gm.gr.InGame() {
			return gameName, nil
		}
		Sleep(1000)
	}

	return gameName, errors.New("error creating game! Timeout")
}

func (gm *GameManager) JoinOnlineGame(gameName, password string) error {
	// Enter bnet lobby
	hid.MovePointer(744, 650)
	hid.Click(hid.LeftButton)
	Sleep(1200)

	// Click "Join game" tab
	hid.MovePointer(977, 54)
	hid.Click(hid.LeftButton)
	Sleep(200)

	// Click the game name textbox, delete text and type new game name
	hid.MovePointer(836, 100)
	hid.Click(hid.LeftButton)
	Sleep(200)
	hid.PressKeyCombination("lctrl", "a")
	Sleep(200)
	for _, ch := range gameName {
		hid.PressKey(fmt.Sprintf("%c", ch))
	}

	// Same for password
	hid.MovePointer(1020, 100)
	hid.Click(hid.LeftButton)
	Sleep(200)
	hid.PressKeyCombination("lctrl", "a")
	Sleep(200)
	for _, ch := range password {
		hid.PressKey(fmt.Sprintf("%c", ch))
	}
	hid.PressKey("enter")

	for i := 0; i < 30; i++ {
		if gm.gr.InGame() {
			return nil
		}
		Sleep(1000)
	}

	return errors.New("error joining game! Timeout")
}
