package config

import (
	"fmt"
	"os"

	"github.com/lxn/win"
	cp "github.com/otiai10/copy"
)

var userProfile = os.Getenv("USERPROFILE")
var settingsPath = userProfile + "\\Saved Games\\Diablo II Resurrected"

func ReplaceGameSettings(modName string) error {
	modDirPath := settingsPath + "\\mods\\" + modName
	modSettingsPath := modDirPath + "\\Settings.json"

	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return fmt.Errorf("game settings not found at %s", settingsPath)
	}

	if _, err := os.Stat(modDirPath); os.IsNotExist(err) {
		err = os.MkdirAll(modDirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating mod folder to store settings: %w", err)
		}
	}

	if _, err := os.Stat(modSettingsPath + ".bkp"); os.IsExist(err) {
		err = os.Rename(modSettingsPath, modSettingsPath+".bkp")
		// File does not exist, no need to back up
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return cp.Copy("config/Settings.json", modSettingsPath)
}

func InstallMod() error {
	if _, err := os.Stat(Koolo.D2RPath + "\\d2r.exe"); os.IsNotExist(err) {
		return fmt.Errorf("game not found at %s", Koolo.D2RPath)
	}

	if _, err := os.Stat(Koolo.D2RPath + "\\mods\\koolo\\koolo.mpq\\modinfo.json"); err == nil {
		return nil
	}

	if err := os.MkdirAll(Koolo.D2RPath+"\\mods\\koolo\\koolo.mpq", os.ModePerm); err != nil {
		return fmt.Errorf("error creating mod folder: %w", err)
	}

	modFileContent := []byte(`{"name":"koolo","savepath":"koolo/"}`)

	return os.WriteFile(Koolo.D2RPath+"\\mods\\koolo\\koolo.mpq\\modinfo.json", modFileContent, 0644)
}

func GetCurrentDisplayScale() float64 {
	hDC := win.GetDC(0)
	defer win.ReleaseDC(0, hDC)
	dpiX := win.GetDeviceCaps(hDC, win.LOGPIXELSX)

	return float64(dpiX) / 96.0
}
