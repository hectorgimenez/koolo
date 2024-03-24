package config

import (
	"encoding/json"
	"fmt"
	"github.com/lxn/win"
	"os"
	"strings"
)

var userProfile = os.Getenv("USERPROFILE")
var settingsFilePath = userProfile + "\\Saved Games\\Diablo II Resurrected\\Settings.json"

func AdjustGameSettings() error {
	settingsJson, err := readGameSettings()
	if err != nil {
		return err
	}

	scale := GetCurrentDisplayScale()
	settingsJson["Screen Resolution (Windowed)"] = fmt.Sprintf("%dx%d", int(1280*scale), int(720*scale))
	settingsJson["Window Mode"] = 0

	settingsContent, err := json.MarshalIndent(settingsJson, "", " ")
	if err != nil {
		return fmt.Errorf("error marshalling settings file: %w", err)
	}

	return os.WriteFile(settingsFilePath, settingsContent, 0666)
}

func AreGameSettingsAdjusted() (bool, error) {
	settingsJson, err := readGameSettings()
	if err != nil {
		return false, err
	}

	windowMode, ok := settingsJson["Window Mode"]
	if !ok || int(windowMode.(float64)) != 0 {
		return false, nil
	}

	scale := GetCurrentDisplayScale()
	targetRes := fmt.Sprintf("%dx%d", int(1280*scale), int(720*scale))

	resolution, ok := settingsJson["Screen Resolution (Windowed)"]
	if !ok || !strings.Contains(resolution.(string), targetRes) {
		return false, nil
	}

	return true, nil
}

func readGameSettings() (map[string]interface{}, error) {
	_, err := os.Stat(settingsFilePath)
	if err != nil {
		return nil, fmt.Errorf("settings file not found: %w", err)
	}

	settingsContent, err := os.ReadFile(settingsFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading settings file: %w", err)
	}

	var settingsJson map[string]interface{}
	err = json.Unmarshal(settingsContent, &settingsJson)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling settings file: %w", err)
	}

	return settingsJson, nil
}

func GetCurrentDisplayScale() float64 {
	hDC := win.GetDC(0)
	defer win.ReleaseDC(0, hDC)
	dpiX := win.GetDeviceCaps(hDC, win.LOGPIXELSX)

	return float64(dpiX) / 96.0
}
