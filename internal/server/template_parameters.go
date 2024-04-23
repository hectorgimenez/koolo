package server

import (
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/config"
)

type IndexData struct {
	ErrorMessage string
	Version      string
	Status       map[string]koolo.Stats
}

type CharacterSettings struct {
	ErrorMessage string
	Supervisor   string
	Config       *config.CharacterCfg
	EnabledRuns  []string
	DisabledRuns []string
}

type ConfigData struct {
	ErrorMessage string
	*config.KooloCfg
}

type AutoSettings struct {
	ErrorMessage string
}
