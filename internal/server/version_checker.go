package server

import (
	"encoding/json"
	"net/http"
	"time"
)

type VersionChecker struct {
	latestVersion string
	latestCheck   time.Time
}

type releaseInfo struct {
	TagName string `json:"tag_name"`
}

func NewVersionChecker() *VersionChecker {
	return &VersionChecker{}
}

func (vc *VersionChecker) Check() string {
	if time.Since(vc.latestCheck) < time.Hour {
		return vc.latestVersion
	}

	r, err := http.Get("https://api.github.com/repos/hectorgimenez/koolo/releases/latest")
	if err != nil {
		return ""
	}
	defer r.Body.Close()

	var ri releaseInfo
	err = json.NewDecoder(r.Body).Decode(&ri)
	if err != nil {
		return ""
	}

	vc.latestVersion = ri.TagName
	vc.latestCheck = time.Now()

	return ri.TagName
}
