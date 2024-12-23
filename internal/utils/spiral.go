package utils

import (
	"math"

	"github.com/hectorgimenez/d2go/pkg/data/entrance"
	"github.com/hectorgimenez/d2go/pkg/data/object"
)

func Spiral(position int) (int, int) {
	t := position * 40
	a := 4.0
	b := -2.0
	trad := float64(t) * math.Pi / 180.0
	x := (a + b*trad) * math.Cos(trad)
	y := (a + b*trad) * math.Sin(trad)

	return int(x), int(y)
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

// EntranceSpiral calculates spiral pattern specifically for entrances/stairs
func EntranceSpiral(attempt int, desc entrance.Description) (x, y int) {
	baseRadius := float64(attempt) * 3.0
	angle := float64(attempt) * math.Pi * (3.0 - math.Sqrt(5.0))

	// Scale based on entrance dimensions
	xScale := float64(desc.SelectDX) / 100.0
	yScale := float64(desc.SelectDY) / 100.0

	// Calculate base position
	x = int(baseRadius * math.Cos(angle) * xScale)
	y = int(baseRadius * math.Sin(angle) * yScale)

	// Apply entrance-specific offsets
	x += desc.SelectX
	y += desc.SelectY

	// Apply direction-specific adjustments
	switch desc.Direction {
	case "l": // left
		x -= 10
	case "r": // right
		x += 10
	}

	// Apply final offsets
	x += desc.OffsetX
	y += desc.OffsetY

	// Clamp values to reasonable ranges based on SelectDX/DY
	maxX := desc.SelectDX / 2
	maxY := desc.SelectDY / 2
	x = Clamp(x, -maxX, maxX)
	y = Clamp(y, -maxY, maxY)

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
