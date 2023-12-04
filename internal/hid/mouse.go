package hid

import (
	asm "github.com/hectorgimenez/koolo/internal/memory"
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

	// TODO: Calculate properly getting desktop scale
	x = int(float32(x) * 1.5)
	y = int(float32(y) * 1.5)

	asm.InjectCursorPos(x, y)
	lParam := calculateLparam(x, y)
	win.SendMessage(HWND, win.WM_NCHITTEST, 0, lParam)             // Set mouse
	win.SendMessage(HWND, win.WM_SETCURSOR, 0x000105A8, 0x2010001) // Set mouse
	win.PostMessage(HWND, win.WM_MOUSEMOVE, 0, lParam)             // Set mouse
	time.Sleep(time.Millisecond * 1)
}

type Changeable interface {
	Set(x, y int, c color.Color)
}

// Click just does a single mouse click at current pointer position
func Click(btn MouseButton, x, y int) {
	MovePointer(x, y)
	x = WindowLeftX + x
	y = WindowTopY + y

	// TODO: Calculate properly getting desktop scale
	x = int(float32(x) * 1.5)
	y = int(float32(y) * 1.5)

	lParam := calculateLparam(x, y)
	buttonDown := uint32(win.WM_LBUTTONDOWN)
	buttonUp := uint32(win.WM_LBUTTONUP)
	if btn == RightButton {
		buttonDown = win.WM_RBUTTONDOWN
		buttonUp = win.WM_RBUTTONUP
	}

	win.SendMessage(HWND, buttonDown, 1, lParam)
	sleepTime := rand.Intn(keyPressMaxTime-keyPressMinTime) + keyPressMinTime
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	win.SendMessage(HWND, buttonUp, 1, lParam)
}

func ClickWithModifier(btn MouseButton, x, y int, modifier ModifierKey) {
	asm.OverrideGetKeyState(int(modifier))
	Click(btn, x, y)
	asm.RestoreGetKeyState()
}

func calculateLparam(x, y int) uintptr {
	return uintptr(y<<16 | x)
}
