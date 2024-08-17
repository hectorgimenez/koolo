package utils

import (
	"math/rand"
	"time"
)

func RandRng(min, max int) int {
	return rand.Intn(max+1-min) + min
}

func RandomDurationMs(min, max int) time.Duration {
	return time.Duration(RandRng(min, max)) * time.Millisecond
}
