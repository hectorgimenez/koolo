package utils

import (
	"math"

	"github.com/hectorgimenez/d2go/pkg/data/entrance"
	"github.com/hectorgimenez/d2go/pkg/data/object"
)

func ItemSpiral(position int) (int, int) {
	t := position * 25

	a := 3.0  // - a controls the starting radius
	b := -1.5 // - b controls how quickly the spiral expands

	// Convert to radians and calculate position
	trad := float64(t) * math.Pi / 180.0

	x := (a + b*trad) * math.Cos(trad)
	y := (a + b*trad) * math.Sin(trad)

	return int(x), int(y)
}

// EntranceSpiral calculates spiral pattern specifically for entrances/stairs
func EntranceSpiral(attempt int, desc entrance.Description) (x, y int) {
	// Calculate base spiral position using golden angle
	angle := float64(attempt) * math.Pi * (3.0 - math.Sqrt(5.0))
	radius := float64(attempt) * 3.0

	// Get entrance dimensions, defaulting to reasonable values if unspecified
	width := 80.0
	height := 100.0
	if desc.SelectDX > 0 {
		width = float64(desc.SelectDX)
	}
	if desc.SelectDY > 0 {
		height = float64(desc.SelectDY)
	}

	// Calculate scaled x,y coordinates
	x = int(radius * math.Cos(angle) * (width / 100.0))
	y = int(radius * math.Sin(angle) * (height / 100.0))

	// Apply entrance offsets
	x += desc.SelectX
	y += desc.SelectY

	// Directional adjustments
	switch desc.Direction {
	case "l":
		x -= 10
	case "r":
		x += 10
	case "b":
		y += 10
	}

	// Clamp values to entrance bounds
	if width > 0 {
		maxX := int(width / 2)
		x = Clamp(x, -maxX, maxX)
	}
	if height > 0 {
		maxY := int(height / 2)
		y = Clamp(y, -maxY, maxY)
	}

	return x, y
}

// ObjectSpiral calculates spiral pattern for objects (portals, chests, etc)
func ObjectSpiral(attempt int, desc object.Description) (x, y int) {
	baseRadius := float64(attempt) * 3.0
	angle := float64(attempt) * math.Pi * (3.0 - math.Sqrt(5.0))

	// Special handling for portals
	if desc.Width == 80 && desc.Height == 110 {
		xScale := 1.0
		yScale := 110.0 / 80.0

		x = int(baseRadius * math.Cos(angle) * xScale)
		y = int(baseRadius*math.Sin(angle)*yScale) - 50

		x = Clamp(x, -40, 40)
		y = Clamp(y, -100, 10)

		return x, y
	}

	// For other objects with dimensions
	if desc.Width > 0 && desc.Height > 0 {
		xScale := float64(desc.Width) / 80.0
		yScale := float64(desc.Height) / 80.0

		x = int(baseRadius * math.Cos(angle) * xScale)
		y = int(baseRadius * math.Sin(angle) * yScale)

		x += desc.Xoffset
		y += desc.Yoffset
		x = Clamp(x, desc.Left, desc.Left+desc.Width)
		y = Clamp(y, desc.Top, desc.Top+desc.Height)

		return x, y
	}

	// Basic object pattern
	x = int(baseRadius * math.Cos(angle))
	y = int(baseRadius * math.Sin(angle))
	return x, y
}

func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
