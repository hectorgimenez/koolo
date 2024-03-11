package server

import (
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/config"
)

type IndexData struct {
	ErrorMessage string
	Status       map[string]koolo.Stats
}

type CharacterSettings struct {
	ErrorMessage string
	IsNew        bool
	Supervisor   string
	*config.CharacterCfg
}

type ConfigData struct {
	ErrorMessage string
	*config.KooloCfg
}
