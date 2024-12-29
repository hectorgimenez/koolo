package utils

import (
	"math"

	"github.com/hectorgimenez/d2go/pkg/data/entrance"
	"github.com/hectorgimenez/d2go/pkg/data/object"
)

func Spiral(position int) (int, int) {
	t := position * 25

	a := 3.0  // - a controls the starting radius
	b := -1.5 // - b controls how quickly the spiral expands

	// Convert to radians and calculate position
	trad := float64(t) * math.Pi / 180.0

	// Calculate spiral coordinates with a slight vertical bias since
	// D2 uses isometric projection (items appear higher than their actual position)
	x := (a + b*trad) * math.Cos(trad)
	y := (a + b*trad) * math.Sin(trad) * 0.9 // Slight vertical compression

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
	// Base radius grows with each attempt, with additional scaling based on the entrance size
	baseRadius := float64(attempt) * 3.0
	// Angle follows the golden angle, which provides a natural spiral pattern
	angle := float64(attempt) * math.Pi * (3.0 - math.Sqrt(5.0))

	// Calculate the scaling based on entrance dimensions
	// We use SelectDX/SelectDY to scale the base radius, so the spiral fits in the entrance bounds
	// If SelectDX or SelectDY is 0, we fall back to a default scaling of 1.0
	xScale := 1.0
	yScale := 1.0

	if desc.SelectDX != 0 {
		xScale = float64(desc.SelectDX) / 100.0
	}
	if desc.SelectDY != 0 {
		yScale = float64(desc.SelectDY) / 100.0
	}

	// Apply scaling to the base radius to compute x and y positions
	// The radius is scaled for each dimension separately, allowing the spiral to fit within the bounds
	x = int(baseRadius * math.Cos(angle) * xScale)
	y = int(baseRadius * math.Sin(angle) * yScale)

	// Apply entrance-specific offsets (shift coordinates based on SelectX and SelectY)
	x += desc.SelectX
	y += desc.SelectY

	// Directional adjustments
	switch desc.Direction {
	case "l": // left
		x -= 10
	case "r": // right
		x += 10
	case "b": // bottom
		y += 10
	}

	// Apply final offsets (OffsetX and OffsetY)
	x += desc.OffsetX
	y += desc.OffsetY

	// Clamp values to the dimensions of the entrance (based on SelectDX and SelectDY)
	// If SelectDX or SelectDY is zero, don't apply clamping to avoid errors
	if desc.SelectDX != 0 && desc.SelectDY != 0 {
		maxX := desc.SelectDX / 2
		maxY := desc.SelectDY / 2
		x = Clamp(x, -maxX, maxX)
		y = Clamp(y, -maxY, maxY)
	}

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
