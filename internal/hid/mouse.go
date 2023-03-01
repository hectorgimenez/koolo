package hid

import (
	"github.com/go-vgo/robotgo"
	"math"
	"math/rand"
)

var (
	WindowLeftX   = 0
	WindowTopY    = 0
	GameAreaSizeX = 0
	GameAreaSizeY = 0
)

const (
	// Windmouse configuration
	windmouseG0 = 9.8  // magnitude of the gravitational force
	windmouseW0 = 5.0  // magnitude of the wind force fluctuations
	windmouseM0 = 1.0  // maximum step size (velocity clip threshold)
	windmouseD0 = 12.0 // distance where wind behavior changes from random to damped

	RightButton MouseButton = "right"
	LeftButton  MouseButton = "left"
)

type MouseButton string

// MovePointer moves the mouse to the requested position, x and y should be the final position based on
// pixels shown in the screen. Top-left corner is 0,0
func MovePointer(x, y int) {
	x = WindowLeftX + x
	y = WindowTopY + y
	M0 := windmouseM0
	destinationX, destinationY := float64(x), float64(y)
	startXi, startYi := getCurrentPosition()
	startX, startY := float64(startXi), float64(startYi)
	currentX, currentY := startX, startY
	Vx, Vy, Wx, Wy := float64(0), float64(0), float64(0), float64(0)

	dist := math.Hypot(destinationX-startX, destinationY-startY)
	for dist >= 1 {
		WMag := math.Min(windmouseW0, dist)

		if dist >= windmouseD0 {
			Wx = Wx/math.Sqrt(3) + (2*rand.Float64()-1)*WMag/math.Sqrt(5)
			Wy = Wy/math.Sqrt(3) + (2*rand.Float64()-1)*WMag/math.Sqrt(5)
		} else {
			Wx /= math.Sqrt(3)
			Wy /= math.Sqrt(3)
			if M0 < 3 {
				M0 = rand.Float64()*3 + 3
			} else {
				M0 /= math.Sqrt(5)
			}
		}

		Vx += Wx + windmouseG0*(destinationX-startX)/dist
		Vy += Wy + windmouseG0*(destinationY-startY)/dist
		VMag := math.Hypot(Vx, Vy)

		if VMag > M0 {
			VClip := M0/2 + rand.Float64()*M0/2
			Vx = (Vx / VMag) * VClip
			Vy = (Vy / VMag) * VClip
		}
		startX += Vx
		startY += Vy

		moveX := int(math.Round(startX))
		moveY := int(math.Round(startY))

		if int(currentX) != moveX || int(currentY) != moveY {
			robotgo.Move(moveX, moveY)
		}

		dist = math.Hypot(destinationX-startX, destinationY-startY)
	}
}

// Click just does a single mouse click at current pointer position
func Click(btn MouseButton) {
	robotgo.Click(string(btn))
}

func getCurrentPosition() (int, int) {
	return robotgo.Location()
}
