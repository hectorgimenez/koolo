package game

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procSendMessageW        = user32.NewProc("SendMessageW")
	procClientToScreen      = user32.NewProc("ClientToScreen")
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
	// First try to exit game as fast as possible, without any check, useful when chickening
	gm.hid.PressKey(win.VK_ESCAPE)
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
		gm.hid.PressKey(win.VK_ESCAPE)
		helper.Sleep(1000)
	}

	return errors.New("error exiting game! Timeout")
}

func (gm *Manager) NewGame() error {
	if gm.gr.InGame() {
		return errors.New("character still in a game")
	}

	for range 30 {
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

	createX := difficultyPosition[config.Characters[gm.supervisorName].Game.Difficulty].X
	createY := difficultyPosition[config.Characters[gm.supervisorName].Game.Difficulty].Y
	gm.hid.Click(LeftButton, 600, 650)
	helper.Sleep(250)
	gm.hid.Click(LeftButton, createX, createY)

	for range 12 {
		if gm.gr.InGame() {
			return nil
		}
		helper.Sleep(500)
	}

	return errors.New("timeout")
}

func (gm *Manager) clearGameNameOrPasswordField() {
	for range 16 {
		gm.hid.PressKey(win.VK_BACK)
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
	gameName := config.Characters[gm.supervisorName].Companion.GameNameTemplate + fmt.Sprintf("%d", gameCounter)
	for _, ch := range gameName {
		gm.hid.PressKey(gm.hid.GetASCIICode(fmt.Sprintf("%c", ch)))
	}

	// Same for password
	gm.hid.Click(LeftButton, 1000, 161)
	helper.Sleep(200)
	gamePassword := config.Characters[gm.supervisorName].Companion.GamePassword
	if gamePassword != "" {
		gm.clearGameNameOrPasswordField()
		for _, ch := range gamePassword {
			gm.hid.PressKey(gm.hid.GetASCIICode(fmt.Sprintf("%c", ch)))
		}
	}
	gm.hid.PressKey(win.VK_RETURN)

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
		gm.hid.PressKey(gm.hid.GetASCIICode(fmt.Sprintf("%c", ch)))
	}

	// Same for password
	gm.hid.Click(LeftButton, 1130, 100)
	helper.Sleep(200)
	gm.clearGameNameOrPasswordField()
	helper.Sleep(200)
	for _, ch := range password {
		gm.hid.PressKey(gm.hid.GetASCIICode(fmt.Sprintf("%c", ch)))
	}
	gm.hid.PressKey(win.VK_RETURN)

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

func terminateProcessByName(name string) error {
	const (
		PROCESS_TERMINATE  = 0x0001
		MAX_PATH           = 260
		TH32CS_SNAPPROCESS = 0x00000002
	)
	hSnapshot, err := windows.CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return fmt.Errorf("failed to create process snapshot")
	}
	defer windows.CloseHandle(hSnapshot)

	var pe32 windows.ProcessEntry32
	pe32.Size = uint32(unsafe.Sizeof(pe32))

	if err := windows.Process32First(hSnapshot, &pe32); err != nil {
		return fmt.Errorf("error during process list")
	}

	var pid uint32
	for {
		processName := windows.UTF16ToString(pe32.ExeFile[:])
		if processName == name {
			pid = pe32.ProcessID
			break
		}
		if err := windows.Process32Next(hSnapshot, &pe32); err != nil {
			if err == syscall.ERROR_NO_MORE_FILES {
				break
			}
		}
	}

	if pid == 0 {
		return nil
	}

	hProcess, err := windows.OpenProcess(PROCESS_TERMINATE, false, pid)
	if err != nil {
		return err
	}

	defer windows.CloseHandle(hProcess)

	if err := windows.TerminateProcess(hProcess, 0); err != nil {
		return fmt.Errorf("failed to terminate process")
	}

	return nil
}

// HELPER FUNCTIONS

func SetForegroundWindow(hwnd windows.HWND) bool {
	ret, _, _ := procSetForegroundWindow.Call(uintptr(hwnd))
	return ret != 0
}

func SendMessage(hwnd windows.HWND, msg uint32, wparam, lparam uintptr) uintptr {
	ret, _, _ := procSendMessageW.Call(
		uintptr(hwnd),
		uintptr(msg),
		wparam,
		lparam,
	)
	return ret
}

type POINT struct {
	X, Y int32
}

func ClientToScreen(hwnd windows.HWND, point *POINT) bool {
	ret, _, _ := procClientToScreen.Call(uintptr(hwnd), uintptr(unsafe.Pointer(point)))
	return ret != 0
}

// END HELPER FUNCTIONS

func StartGame(username string, password string, authmethod string, realm string, arguments string, useCustomSettings bool) (uint32, win.HWND, error) {
	// First check for other instances of the game and kill the handles, otherwise we will not be able to start the game
	err := KillAllClientHandles()
	if err != nil {
		return 0, 0, err
	}

	// Depending on the authentication method set base arguments
	var baseArgs []string

	if authmethod == "BattleNetClient" {
		baseArgs = []string{"-uid osi"}
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

	// Add them to the full argument list
	fullArgs := append(baseArgs, additionalArguments...)

	var bnetCmd *exec.Cmd

	// If auth method is set to battlenet client, start the process
	if authmethod == "BattleNetClient" {

		// First we check if the process exists, if so, terminate it
		// We're looking for the window by title as there can be many Battle.net.exe processes
		err := terminateProcessByName("Battle.net.exe")
		if err != nil {
			return 0, 0, err
		}

		// Start the Battle.net Process
		bnetCmd = exec.Command("C:\\Program Files (x86)\\Battle.net\\Battle.net.exe", "--from-launcher")
		err = bnetCmd.Start()
		if err != nil {
			return 0, 0, err
		}

		// Give enough time for the process to start
		helper.Sleep(5000)

		// Log in process
		var bnetHandle windows.HWND

		cb := syscall.NewCallback(func(hwnd windows.HWND, lParam uintptr) uintptr {
			var pid uint32
			windows.GetWindowThreadProcessId(hwnd, &pid)
			if pid == uint32(bnetCmd.Process.Pid) {
				bnetHandle = hwnd
				return 0
			}
			return 1
		})
		for {
			windows.EnumWindows(cb, unsafe.Pointer(&bnetCmd.Process.Pid))
			if bnetHandle != 0 {
				// Small delay and read again, to be sure we are capturing the right hwnd
				time.Sleep(time.Second)
				windows.EnumWindows(cb, unsafe.Pointer(&bnetCmd.Process.Pid))
				break
			}
		}

		const (
			WM_LBUTTONDOWN = 0x0201
			WM_LBUTTONUP   = 0x0202
			WM_KEYDOWN     = 0x0100
			WM_KEYUP       = 0x0101
			VK_CONTROL     = 0x11
			VK_A           = 0x41
			VK_BACK        = 0x08
			MK_LBUTTON     = 0x0001
			VK_TAB         = 0x09
		)

		if bnetHandle == 0 {
			return 0, 0, errors.New("failed to find Battle.net handle")
		}

		helper.Sleep(5000)

		// Bring the window to front
		SetForegroundWindow(bnetHandle)

		x, y := int32(216), int32(275)

		// Convert client coordinates to screen coordinates
		point := POINT{X: x, Y: y}
		if !ClientToScreen(bnetHandle, &point) {
			fmt.Println("Failed to convert client coordinates to screen coordinates")
		}

		lparam := uintptr((point.Y << 16) | (point.X & 0xFFFF))

		SendMessage(bnetHandle, WM_LBUTTONDOWN, MK_LBUTTON, lparam)
		SendMessage(bnetHandle, WM_LBUTTONUP, 0, lparam)
		helper.Sleep(100)
		SendMessage(bnetHandle, WM_LBUTTONDOWN, MK_LBUTTON, lparam)
		SendMessage(bnetHandle, WM_LBUTTONUP, 0, lparam)
		helper.Sleep(250)
		SendMessage(bnetHandle, WM_KEYDOWN, VK_BACK, 0)
		SendMessage(bnetHandle, WM_KEYUP, VK_BACK, 0)

		helper.ShowDialog("Username should be deleted", "Username should be deleted")
		os.Exit(0)

		// Type out the username
		for _, char := range username {
			SendMessage(bnetHandle, win.WM_CHAR, uintptr(char), 0)
			log.Print("Posted WM_CHAR message")
			helper.Sleep(50)
		}

		// Click on the password field
		SendMessage(bnetHandle, WM_LBUTTONDOWN, MK_LBUTTON, calculateLparam2(209, 329))
		log.Printf("Posted WM_LBUTTONDOWN MK_LBUTTON message")
		helper.Sleep(100)
		SendMessage(bnetHandle, WM_LBUTTONUP, MK_LBUTTON, calculateLparam2(209, 329))
		log.Printf("Posted WM_LBUTTONUP MK_LBUTTON message")
		helper.Sleep(100)

		// Delete the current text that's in the password field
		SendMessage(bnetHandle, WM_KEYDOWN, VK_CONTROL, 0)
		log.Printf("Sent VK_CONTROL WM_KEYDOWN message")
		SendMessage(bnetHandle, WM_KEYDOWN, VK_A, 0)
		log.Printf("Sent WM_KEYDOWN VK_A message")
		SendMessage(bnetHandle, WM_KEYUP, VK_A, 0)
		log.Printf("Sent WM_KEYUP VK_A message")
		SendMessage(bnetHandle, WM_KEYUP, VK_CONTROL, 0)
		log.Printf("Sent VK_CONTROL WM_KEYUP message")
		helper.Sleep(1000)
		SendMessage(bnetHandle, WM_KEYDOWN, VK_BACK, 0)
		log.Printf("Sent VK_BACK WM_KEYDOWN message")
		SendMessage(bnetHandle, WM_KEYUP, VK_BACK, 0)
		log.Printf("Sent VK_BACK WM_KEYUP message")
		helper.Sleep(100)

		// Type out the password
		for _, char := range password {
			SendMessage(bnetHandle, win.WM_CHAR, uintptr(char), 0)
			log.Print("Sent WM_CHAR message")
			helper.Sleep(50)
		}

		// Click on the login button
		SendMessage(bnetHandle, WM_LBUTTONDOWN, MK_LBUTTON, calculateLparam2(222, 451))
		log.Printf("Sent WM_LBUTTONDOWN MK_LBUTTON message")
		SendMessage(bnetHandle, WM_LBUTTONUP, MK_LBUTTON, calculateLparam2(222, 451))
		log.Printf("Sent WM_LBUTTONUP MK_LBUTTON message")

		// Wait for the login to finish
		helper.Sleep(5000)
	}

	// Start the game
	cmd := exec.Command(config.Koolo.D2RPath+"\\D2R.exe", fullArgs...)

	if useCustomSettings {
		err = config.ReplaceGameSettings()
		if err != nil {
			return 0, 0, err
		}
	}

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
