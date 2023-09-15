package hid

import (
	"github.com/go-vgo/robotgo"
)

var (
	WindowLeftX   = 0
	WindowTopY    = 0
	GameAreaSizeX = 0
	GameAreaSizeY = 0
)

const (
	RightButton MouseButton = "right"
	LeftButton  MouseButton = "left"
)

type MouseButton string

// MovePointer moves the mouse to the requested position, x and y should be the final position based on
// pixels shown in the screen. Top-left corner is 0,0
func MovePointer(x, y int) {
	x = WindowLeftX + x
	y = WindowTopY + y

	// TODO this is a workaround to prevent moving the mouse outside the game window
	if x > WindowLeftX+GameAreaSizeX {
		x = WindowLeftX + GameAreaSizeX - 10
	}
	if y > WindowTopY+GameAreaSizeY {
		y = WindowTopY + GameAreaSizeY - 10
	}

	robotgo.Move(x, y)
}

// Click just does a single mouse click at current pointer position
func Click(btn MouseButton) {
	robotgo.Click(string(btn))
}
