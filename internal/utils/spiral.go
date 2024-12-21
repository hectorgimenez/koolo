package utils

import (
	"math"

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

func AdaptiveSpiral(attempt int, desc object.Description) (x, y int) {
	baseRadius := float64(attempt) * 3.0
	angle := float64(attempt) * math.Pi * (3.0 - math.Sqrt(5.0))

	// If object has no dimensions (like entrances), use entrance-specific pattern
	if desc.Width == 0 && desc.Height == 0 {
		// Detect cellar/tower/underground entrance by checking the Top offset
		// These entrances typically have a higher negative Top value
		isTallEntrance := desc.Top < -45

		xScale := 1.2
		yScale := 0.8
		yOffset := -35

		if isTallEntrance {
			// Make the pattern taller and narrower for vertical entrances
			xScale = 0.9
			yScale = 1.1
			yOffset = -45 // Search higher up
		}

		x = int(baseRadius * math.Cos(angle) * xScale)
		y = int(baseRadius * math.Sin(angle) * yScale)

		y += yOffset

		// Adjust boundaries for vertical entrances
		if isTallEntrance {
			x = Clamp(x, -25, 25) // Narrower X range
			y = Clamp(y, -60, 0)  // Higher Y range
		} else {
			x = Clamp(x, -30, 30)
			y = Clamp(y, -50, 10)
		}

		return x, y
	}
	// For portal-like objects (similar dimensions)
	if desc.Width == 80 && desc.Height == 110 {
		xScale := 1.0
		yScale := 110.0 / 80.0

		x = int(baseRadius * math.Cos(angle) * xScale)
		y = int(baseRadius*math.Sin(angle)*yScale) - 50

		x = Clamp(x, -40, 40)
		y = Clamp(y, -100, 10)

		return x, y
	}

	// For other objects with dimensions, scale based on their actual size
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
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
