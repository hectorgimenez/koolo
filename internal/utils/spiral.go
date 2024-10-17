package utils

import "math"

func Spiral(position int) (int, int) {
	t := position * 40
	a := 4.0
	b := -2.0
	trad := float64(t) * math.Pi / 180.0
	x := (a + b*trad) * math.Cos(trad)
	y := (a + b*trad) * math.Sin(trad)

	return int(x), int(y)
}
