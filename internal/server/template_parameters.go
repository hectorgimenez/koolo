package server

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/v2/bot"
)

type IndexData struct {
	ErrorMessage string
	Version      string
	Status       map[string]bot.Stats
	DropCount    map[string]int
}

type DropData struct {
	NumberOfDrops int
	Character     string
	Drops         []data.Drop
}

type CharacterSettings struct {
	ErrorMessage string
	Supervisor   string
	Config       *config.CharacterCfg
	EnabledRuns  []string
	DisabledRuns []string
	AvailableTZs map[int]string
	RecipeList   []string
}

type ConfigData struct {
	ErrorMessage string
	*config.KooloCfg
}

type AutoSettings struct {
	ErrorMessage string
}
