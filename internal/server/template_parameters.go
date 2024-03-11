package server

import (
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/config"
)

type IndexData struct {
	Status map[string]koolo.Stats
}

type CharacterSettings struct {
	IsNew      bool
	Supervisor string
	*config.CharacterCfg
}
