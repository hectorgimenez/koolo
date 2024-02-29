package game

import (
	"errors"
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type Manager struct {
	gr  *MemoryReader
	hid *HID
}

func NewGameManager(gr *MemoryReader, hid *HID) *Manager {
	return &Manager{gr: gr, hid: hid}
}

func (gm *Manager) ExitGame() error {
	// First try to exit game as fast as possible, without any check, useful when chickening
	gm.hid.PressKey("esc")
	gm.hid.Click(LeftButton, gm.gr.GameAreaSizeX/2, int(float64(gm.gr.GameAreaSizeY)/2.2))

	for range 5 {
		if !gm.gr.InGame() {
			return nil
		}
		helper.Sleep(1000)
	}

	// If we are still in game, probably character is dead, so let's do it nicely.
	// Probably closing the socket is more reliable, but was not working properly for me on singleplayer.
	for range 10 {
		if gm.gr.GetData(false).OpenMenus.QuitMenu {
			gm.hid.Click(LeftButton, gm.gr.GameAreaSizeX/2, int(float64(gm.gr.GameAreaSizeY)/2.2))

			for range 5 {
				if !gm.gr.InGame() {
					return nil
				}
				helper.Sleep(1000)
			}
		}
		gm.hid.PressKey("esc")
		helper.Sleep(1000)
	}

	return errors.New("error exiting game! Timeout")
}

func (gm *Manager) NewGame() error {
	if gm.gr.InGame() {
		return errors.New("character still in a game")
	}

	for range 30 {
		gm.gr.InGame()
		if gm.gr.InCharacterSelectionScreen() {
			helper.Sleep(2000) // Wait for character selection screen to load
			break
		}
		helper.Sleep(500)
	}

	difficultyPosition := map[difficulty.Difficulty]struct {
		X, Y int
	}{
		difficulty.Normal:    {X: 640, Y: 311},
		difficulty.Nightmare: {X: 640, Y: 355},
		difficulty.Hell:      {X: 640, Y: 403},
	}

	createX := difficultyPosition[config.Config.Game.Difficulty].X
	createY := difficultyPosition[config.Config.Game.Difficulty].Y
	gm.hid.Click(LeftButton, 600, 650)
	helper.Sleep(250)
	gm.hid.Click(LeftButton, createX, createY)

	for range 30 {
		if gm.gr.InGame() {
			return nil
		}
		helper.Sleep(1000)
	}

	return errors.New("error creating game! Timeout")
}

func (gm *Manager) clearGameNameOrPasswordField() {
	for range 16 {
		gm.hid.PressKey("backspace")
	}
}

func (gm *Manager) CreateOnlineGame(gameCounter int) (string, error) {
	// Enter bnet lobby
	gm.hid.Click(LeftButton, 744, 650)
	helper.Sleep(1200)

	// Click "Create game" tab
	gm.hid.Click(LeftButton, 845, 54)
	helper.Sleep(200)

	// Click the game name textbox, delete text and type new game name
	gm.hid.Click(LeftButton, 1000, 116)
	gm.clearGameNameOrPasswordField()
	gameName := config.Config.Companion.GameNameTemplate + fmt.Sprintf("%d", gameCounter)
	for _, ch := range gameName {
		gm.hid.PressKey(fmt.Sprintf("%c", ch))
	}

	// Same for password
	gm.hid.Click(LeftButton, 1000, 161)
	helper.Sleep(200)
	gamePassword := config.Config.Companion.GamePassword
	if gamePassword != "" {
		gm.clearGameNameOrPasswordField()
		for _, ch := range gamePassword {
			gm.hid.PressKey(fmt.Sprintf("%c", ch))
		}
	}
	gm.hid.PressKey("enter")

	for range 30 {
		if gm.gr.InGame() {
			return gameName, nil
		}
		helper.Sleep(1000)
	}

	return gameName, errors.New("error creating game! Timeout")
}

func (gm *Manager) JoinOnlineGame(gameName, password string) error {
	// Enter bnet lobby
	gm.hid.Click(LeftButton, 744, 650)
	helper.Sleep(1200)

	// Click "Join game" tab
	gm.hid.Click(LeftButton, 977, 54)
	helper.Sleep(200)

	// Click the game name textbox, delete text and type new game name
	gm.hid.Click(LeftButton, 950, 100)
	helper.Sleep(200)
	gm.clearGameNameOrPasswordField()
	helper.Sleep(200)
	for _, ch := range gameName {
		gm.hid.PressKey(fmt.Sprintf("%c", ch))
	}

	// Same for password
	gm.hid.Click(LeftButton, 1130, 100)
	helper.Sleep(200)
	gm.clearGameNameOrPasswordField()
	helper.Sleep(200)
	for _, ch := range password {
		gm.hid.PressKey(fmt.Sprintf("%c", ch))
	}
	gm.hid.PressKey("enter")

	for range 30 {
		if gm.gr.InGame() {
			return nil
		}
		helper.Sleep(1000)
	}

	return errors.New("error joining game! Timeout")
}

func (gm *Manager) InGame() bool {
	return gm.gr.InGame()
}
