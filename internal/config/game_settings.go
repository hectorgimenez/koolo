package config

import (
	"github.com/lxn/win"
	cp "github.com/otiai10/copy"
	"os"
)

var userProfile = os.Getenv("USERPROFILE")
var settingsFilePath = userProfile + "\\Saved Games\\Diablo II Resurrected\\Settings.json"

func ReplaceGameSettings() error {
	if _, err := os.Stat(settingsFilePath + ".bkp"); os.IsNotExist(err) {
		err = os.Rename(settingsFilePath, settingsFilePath+".bkp")
		// File does not exist, no need to back up
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return cp.Copy("config/Settings.json", settingsFilePath)
}

func GetCurrentDisplayScale() float64 {
	hDC := win.GetDC(0)
	defer win.ReleaseDC(0, hDC)
	dpiX := win.GetDeviceCaps(hDC, win.LOGPIXELSX)

	return float64(dpiX) / 96.0
}
