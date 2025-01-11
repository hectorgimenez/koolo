package utils

import (
	"math"
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

func ItemPickupPattern(attempt int) (int, int) {
	// Uses a grid-based pattern instead of a mathematical spiral for more reliable pickups
	// The pattern starts at center and works outward in a cross/diamond pattern
	patterns := [][2]int{
		{0, 0},   // Center
		{0, -1},  // Up
		{1, 0},   // Right
		{0, 1},   // Down
		{-1, 0},  // Left
		{1, -1},  // Upper right
		{1, 1},   // Lower right
		{-1, 1},  // Lower left
		{-1, -1}, // Upper left
		{0, -2},  // Outer up
		{2, 0},   // Outer right
		{0, 2},   // Outer down
		{-2, 0},  // Outer left
		{1, -2},  // Outer upper right
		{2, -1},
		{2, 1},  // Outer right side
		{1, 2},  // Outer lower right
		{-1, 2}, // Outer lower left
		{-2, 1},
		{-2, -1}, // Outer left side
		{-1, -2}, // Outer upper left
	}

	// If we somehow exceed our pattern, start doing slightly larger offsets
	// This is a fallback for edge cases
	if attempt >= len(patterns) {
		extra := attempt - len(patterns) + 1
		return patterns[attempt%len(patterns)][0] * extra,
			patterns[attempt%len(patterns)][1] * extra
	}

	return patterns[attempt][0], patterns[attempt][1]
}
