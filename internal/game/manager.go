package game

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/billgraziano/dpapi"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

type Manager struct {
	gr             *MemoryReader
	hid            *HID
	supervisorName string
}

func NewGameManager(gr *MemoryReader, hid *HID, sueprvisorName string) *Manager {
	return &Manager{gr: gr, hid: hid, supervisorName: sueprvisorName}
}

func (gm *Manager) ExitGame() error {
	if !gm.gr.InGame() {
		return nil
	}
	// First try to exit game as fast as possible, without any check, useful when chickening
	gm.hid.PressKey(win.VK_ESCAPE)
	gm.hid.Click(LeftButton, gm.gr.GameAreaSizeX/2, int(float64(gm.gr.GameAreaSizeY)/2.2))

	for range 5 {
		if !gm.gr.InGame() {
			return nil
		}
		utils.Sleep(1000)
	}

	// If we are still in game, probably character is dead, so let's do it nicely.
	// Probably closing the socket is more reliable, but was not working properly for me on singleplayer.
	for range 10 {
		if gm.gr.GetData().OpenMenus.QuitMenu {
			gm.hid.Click(LeftButton, gm.gr.GameAreaSizeX/2, int(float64(gm.gr.GameAreaSizeY)/2.2))

			for range 5 {
				if !gm.gr.InGame() {
					return nil
				}
				utils.Sleep(1000)
			}
		}
		gm.hid.PressKey(win.VK_ESCAPE)
		utils.Sleep(1000)
	}

	return errors.New("error exiting game! Timeout")
}

func (gm *Manager) NewGame() error {
	if gm.gr.InGame() {
		return errors.New("character still in a game")
	}

	for range 30 {
		if gm.gr.IsInCharacterSelectionScreen() {
			break
		} else {
			utils.Sleep(500)
		}
	}

	difficultyPosition := map[difficulty.Difficulty]struct {
		X, Y int
	}{
		difficulty.Normal:    {X: 640, Y: 311},
		difficulty.Nightmare: {X: 640, Y: 355},
		difficulty.Hell:      {X: 640, Y: 403},
	}

	createX := difficultyPosition[config.Characters[gm.supervisorName].Game.Difficulty].X
	createY := difficultyPosition[config.Characters[gm.supervisorName].Game.Difficulty].Y
	gm.hid.Click(LeftButton, 600, 650)
	utils.Sleep(250)
	gm.hid.Click(LeftButton, createX, createY)

	for range 12 {
		if gm.gr.InGame() {
			return nil
		}
		utils.Sleep(500)
	}

	return errors.New("timeout")
}

func (gm *Manager) clearGameNameOrPasswordField() {
	for range 16 {
		gm.hid.PressKey(win.VK_BACK)
	}
}

func (gm *Manager) CreateLobbyGame(gameCounter int) (string, error) {

	// Click "Create game" tab
	gm.hid.Click(LeftButton, 845, 54)
	utils.Sleep(200)

	difficultyPosition := map[difficulty.Difficulty]struct {
		X, Y int
	}{
		difficulty.Normal:    {X: 900, Y: 252},
		difficulty.Nightmare: {X: 980, Y: 252},
		difficulty.Hell:      {X: 1065, Y: 252},
	}

	difficultyPos := difficultyPosition[config.Characters[gm.supervisorName].Game.Difficulty]
	gm.hid.Click(LeftButton, difficultyPos.X, difficultyPos.Y)
	utils.Sleep(200)

	// Click the game name textbox, delete text and type new game name
	gm.hid.Click(LeftButton, 1000, 116)
	gm.clearGameNameOrPasswordField()
	gameName := config.Characters[gm.supervisorName].Companion.GameNameTemplate + fmt.Sprintf("%d", gameCounter)
	for _, ch := range gameName {
		gm.hid.PressKey(gm.hid.GetASCIICode(fmt.Sprintf("%c", ch)))
	}

	// Same for password
	gm.hid.Click(LeftButton, 1000, 161)
	utils.Sleep(200)
	gamePassword := config.Characters[gm.supervisorName].Companion.GamePassword
	if gamePassword != "" {
		gm.clearGameNameOrPasswordField()
		for _, ch := range gamePassword {
			gm.hid.PressKey(gm.hid.GetASCIICode(fmt.Sprintf("%c", ch)))
		}
	}
	gm.hid.PressKey(win.VK_RETURN)

	for range 15 {
		if gm.gr.InGame() {
			return gameName, nil
		}
		utils.Sleep(1000)

		panel := gm.gr.GetPanel("DismissableModal")
		if panel.PanelName != "" && panel.PanelEnabled && panel.PanelVisible {
			gm.hid.PressKey(win.VK_ESCAPE)
			utils.Sleep(1000)
			return gameName, errors.New("error creating game! Got error message")
		}
	}

	return gameName, errors.New("error creating game! Timeout")
}

func (gm *Manager) JoinOnlineGame(gameName, password string) error {

	// Click "Join game" tab
	gm.hid.Click(LeftButton, 977, 54)
	utils.Sleep(200)

	// Click the game name textbox, delete text and type new game name
	gm.hid.Click(LeftButton, 950, 100)
	utils.Sleep(200)
	gm.clearGameNameOrPasswordField()
	utils.Sleep(200)
	for _, ch := range gameName {
		gm.hid.PressKey(gm.hid.GetASCIICode(fmt.Sprintf("%c", ch)))
	}

	// Same for password
	gm.hid.Click(LeftButton, 1130, 100)
	utils.Sleep(200)
	gm.clearGameNameOrPasswordField()
	utils.Sleep(200)
	for _, ch := range password {
		gm.hid.PressKey(gm.hid.GetASCIICode(fmt.Sprintf("%c", ch)))
	}
	gm.hid.PressKey(win.VK_RETURN)

	for range 15 {
		if gm.gr.InGame() {
			return nil
		}
		utils.Sleep(1000)

		// Check if we got an error message while trying to join the game
		panel := gm.gr.GetPanel("DismissableModal")
		if panel.PanelName != "" && panel.PanelEnabled && panel.PanelVisible {
			gm.hid.PressKey(win.VK_ESCAPE)
			utils.Sleep(1000)
			return errors.New("error joining game! Got error message")
		}
	}

	return errors.New("error joining game! Timeout")
}

func (gm *Manager) InGame() bool {
	return gm.gr.InGame()
}

func StartGame(username string, password string, authmethod string, authToken string, realm string, arguments string, useCustomSettings bool) (uint32, win.HWND, error) {
	// First check for other instances of the game and kill the handles, otherwise we will not be able to start the game
	err := KillAllClientHandles()
	if err != nil {
		return 0, 0, err
	}

	// Depending on the authentication method set base arguments
	var baseArgs []string

	if authmethod == "TokenAuth" {
		baseArgs = []string{"-uid", "osi"}
	} else if authmethod == "UsernamePassword" {
		baseArgs = []string{"-username", username, "-password", password, "-address", realm}
	} else if authmethod == "None" {
		baseArgs = []string{}
	} else {
		// Default to no auth method
		baseArgs = []string{}
	}

	// Parse the provided additional arguments
	additionalArguments := strings.Fields(arguments)

	// Let's use the mod directory for storing the settings, so we stop overwriting the default config
	if useCustomSettings {
		modName := "koolo"
		found := false
		for i, arg := range additionalArguments {
			if arg == "-mod" {
				modName = additionalArguments[i+1]
				found = true
				break
			}
		}
		if !found {
			additionalArguments = append(additionalArguments, "-mod", modName)
		}

		// If there is no real mod, let's create a fake mod called "koolo" so we can store our own config
		if modName == "koolo" {
			err = config.InstallMod()
			if err != nil {
				return 0, 0, err
			}
		}

		// Replace game mod settings with the custom ones
		err = config.ReplaceGameSettings(modName)
		if err != nil {
			return 0, 0, err
		}
	}

	// Add them to the full argument list
	fullArgs := append(baseArgs, additionalArguments...)

	if authmethod == "TokenAuth" {
		// Entropy buffer
		entropy := []byte{0xc8, 0x76, 0xf4, 0xae, 0x4c, 0x95, 0x2e, 0xfe, 0xf2, 0xfa, 0x0f, 0x54, 0x19, 0xc0, 0x9c, 0x43}
		tokenBytes := []byte(authToken)

		encryptedToken, err := dpapi.EncryptBytesEntropy(tokenBytes, entropy)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to encrypt auth token: %v", err)
		}

		// Create or Open the OSI registry folder
		key, _, err := registry.CreateKey(registry.CURRENT_USER, `SOFTWARE\Blizzard Entertainment\Battle.net\Launch Options\OSI`, registry.ALL_ACCESS)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to open registry key: %v", err)
		}
		defer key.Close()

		region := "EU"
		switch realm {
		case "eu.actual.battle.net":
			region = "EU"
		case "us.actual.battle.net":
			region = "US"
		case "kr.actual.battle.net":
			region = "KR"
		default:
			region = "EU"
		}

		// Update the region registry
		err = key.SetStringValue("REGION", region)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to set REGION registry value: %v", err)
		}

		err = key.SetBinaryValue("WEB_TOKEN", encryptedToken)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to set WEB_TOKEN registry value: %v", err)
		}

		// If we got to here we've successfully updated the auth token :)
	}

	// Start the game
	cmd := exec.Command(config.Koolo.D2RPath+"\\D2R.exe", fullArgs...)
	err = cmd.Start()
	if err != nil {
		return 0, 0, err
	}

	var foundHwnd windows.HWND
	cb := syscall.NewCallback(func(hwnd windows.HWND, lParam uintptr) uintptr {
		var pid uint32
		windows.GetWindowThreadProcessId(hwnd, &pid)
		if pid == uint32(cmd.Process.Pid) {
			foundHwnd = hwnd
			return 0
		}
		return 1
	})
	for {
		windows.EnumWindows(cb, unsafe.Pointer(&cmd.Process.Pid))
		if foundHwnd != 0 {
			// Small delay and read again, to be sure we are capturing the right hwnd
			time.Sleep(time.Second)
			windows.EnumWindows(cb, unsafe.Pointer(&cmd.Process.Pid))
			break
		}
	}

	// Close the handle for the new process, it will allow the user to open another instance of the game
	err = KillAllClientHandles()
	if err != nil {
		return 0, 0, err
	}

	return uint32(cmd.Process.Pid), win.HWND(foundHwnd), nil
}
