package server

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/bot"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
)

type SupervisorCard struct {
	bot.Stats
	Statistics []RunStatistics
	SkillID    skill.ID
	Name       string
}

type RunStatistics struct {
	Name     string
	Fastest  time.Duration
	Average  time.Duration
	Errors   int
	Deaths   int
	Chickens int
	Drops    int
	Slowest  time.Duration
	Total    int
}

type DropPage struct {
	Drops []Drop
}

type Drop struct {
	game.Drop
	Supervisor string
	Run        string
}

type IndexData struct {
	ErrorMessage       string
	Version            string
	StartedSupervisors []SupervisorCard
	StoppedSupervisors []SupervisorCard
}

type CharacterSettings struct {
	ErrorMessage string
	Supervisor   string
	Config       *config.CharacterCfg
	DayNames     []string
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
