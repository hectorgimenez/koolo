package helper

import (
	"math/rand"
	"time"
)

// Sleep provides a Sleep function that randomize the sleep time up/down to a maximum of 30%
func Sleep(milliseconds int) {
	maxTime := int(float32(milliseconds) * 1.3)
	minTime := int(float32(milliseconds) * 0.7)
	sleepTime := rand.Intn(maxTime-minTime) + minTime

	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
}
