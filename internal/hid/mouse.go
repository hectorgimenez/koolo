package hid

import (
	"github.com/hectorgimenez/koolo/internal/memory"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/lxn/win"
	"image/color"
	"math/rand"
	"time"
)

var (
	WindowLeftX   = 0
	WindowTopY    = 0
	GameAreaSizeX = 0
	GameAreaSizeY = 0
)

const (
	RightButton MouseButton = win.MK_RBUTTON
	LeftButton  MouseButton = win.MK_LBUTTON

	ShiftKey ModifierKey = win.VK_SHIFT
	CtrlKey  ModifierKey = win.VK_CONTROL
)

type MouseButton uint
type ModifierKey uint

// MovePointer moves the mouse to the requested position, x and y should be the final position based on
// pixels shown in the screen. Top-left corner is 0,0
func MovePointer(x, y int) {
	x = WindowLeftX + x
	y = WindowTopY + y

	scale := ui.GameWindowScale()
	x = int(float64(x) * scale)
	y = int(float64(y) * scale)

	memory.InjectCursorPos(x, y)
	lParam := calculateLparam(x, y)
	win.SendMessage(memory.HWND, win.WM_NCHITTEST+1000, 0, lParam)
	win.SendMessage(memory.HWND, win.WM_SETCURSOR+1000, 0x000105A8, 0x2010001)
	win.PostMessage(memory.HWND, win.WM_MOUSEMOVE+1000, 0, lParam)
}

type Changeable interface {
	Set(x, y int, c color.Color)
}

// Click just does a single mouse click at current pointer position
func Click(btn MouseButton, x, y int) {
	MovePointer(x, y)
	x = WindowLeftX + x
	y = WindowTopY + y

	scale := ui.GameWindowScale()
	x = int(float64(x) * scale)
	y = int(float64(y) * scale)

	lParam := calculateLparam(x, y)
	buttonDown := uint32(win.WM_LBUTTONDOWN + 1000)
	buttonUp := uint32(win.WM_LBUTTONUP + 1000)
	if btn == RightButton {
		buttonDown = win.WM_RBUTTONDOWN + 1000
		buttonUp = win.WM_RBUTTONUP + 1000
	}

	win.SendMessage(memory.HWND, buttonDown, 1, lParam)
	sleepTime := rand.Intn(keyPressMaxTime-keyPressMinTime) + keyPressMinTime
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	win.SendMessage(memory.HWND, buttonUp, 1, lParam)
}

func ClickWithModifier(btn MouseButton, x, y int, modifier ModifierKey) {
	memory.OverrideGetKeyState(int(modifier))
	Click(btn, x, y)
	memory.RestoreGetKeyState()
}

func calculateLparam(x, y int) uintptr {
	return uintptr(y<<16 | x)
}
