package helper

import (
	"math/rand"
	"time"
)

func RandRng(min, max int) int {
	rand.Seed(time.Now().UnixNano())

	return rand.Intn(max+1-min) + min
}
