package helper

import (
	"fmt"
	"github.com/go-vgo/robotgo"
	"os"
	"time"
)

func Screenshot() {
	if _, err := os.Stat("screenshots"); os.IsNotExist(err) {
		err = os.MkdirAll("screenshots", 0700)
		if err != nil {
			return
		}
	}

	fileName := fmt.Sprintf("screenshots/error-%s.png", time.Now().Format("2006-01-02 15_04_05"))
	robotgo.SaveCapture(fileName)
}
