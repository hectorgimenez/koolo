package context

import (
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/pather"
)

var mu sync.Mutex
var botContexts = make(map[uint64]*Status)

type Priority int

const (
	PriorityHigh       = 0
	PriorityNormal     = 1
	PriorityBackground = 5
	PriorityPause      = 10
	PriorityStop       = 100
)

type Status struct {
	*Context
	Priority Priority
}

type Context struct {
	Name              string
	ExecutionPriority Priority
	CharacterCfg      *config.CharacterCfg
	Data              *game.Data
	EventListener     *event.Listener
	HID               *game.HID
	Logger            *slog.Logger
	Manager           *game.Manager
	GameReader        *game.MemoryReader
	MemoryInjector    *game.MemoryInjector
	PathFinder        *pather.PathFinder
	BeltManager       *health.BeltManager
	HealthManager     *health.Manager
	Char              Character
	LastBuffAt        time.Time
	ContextDebug      map[Priority]*Debug
	CurrentGame       *CurrentGameHelper
}

type Debug struct {
	LastAction string `json:"lastAction"`
	LastStep   string `json:"lastStep"`
}

type CurrentGameHelper struct {
	BlacklistedItems []data.Item
	AreaCorrection   struct {
		Enabled      bool
		ExpectedArea area.ID
	}
	PickupItems bool
}

func NewContext(name string) *Status {
	ctx := &Context{
		Name:              name,
		Data:              &game.Data{},
		ExecutionPriority: PriorityNormal,
		ContextDebug: map[Priority]*Debug{
			PriorityBackground: {},
			PriorityNormal:     {},
			PriorityHigh:       {},
			PriorityPause:      {},
			PriorityStop:       {},
		},
		CurrentGame: &CurrentGameHelper{},
	}
	botContexts[getGoroutineID()] = &Status{Priority: PriorityNormal, Context: ctx}

	return botContexts[getGoroutineID()]
}

func NewGameHelper() *CurrentGameHelper {
	return &CurrentGameHelper{
		PickupItems: true,
	}
}

func Get() *Status {
	mu.Lock()
	defer mu.Unlock()
	return botContexts[getGoroutineID()]
}

func (s *Status) SetLastAction(actionName string) {
	s.Context.ContextDebug[s.Priority].LastAction = actionName
}

func (s *Status) SetLastStep(stepName string) {
	s.Context.ContextDebug[s.Priority].LastStep = stepName
}

func getGoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	stackTrace := string(buf[:n])
	fields := strings.Fields(stackTrace)
	id, _ := strconv.ParseUint(fields[1], 10, 64)

	return id
}

func (ctx *Context) RefreshGameData() {
	*ctx.Data = ctx.GameReader.GetData()
}

func (ctx *Context) Detach() {
	mu.Lock()
	defer mu.Unlock()
	delete(botContexts, getGoroutineID())
}

func (ctx *Context) AttachRoutine(priority Priority) {
	mu.Lock()
	defer mu.Unlock()
	botContexts[getGoroutineID()] = &Status{Priority: priority, Context: ctx}
}

func (ctx *Context) SwitchPriority(priority Priority) {
	ctx.ExecutionPriority = priority
}

func (ctx *Context) DisableItemPickup() {
	ctx.CurrentGame.PickupItems = false
}

func (ctx *Context) EnableItemPickup() {
	ctx.CurrentGame.PickupItems = true
}

func (s *Status) PauseIfNotPriority() {
	// This prevents bot from trying to move when loading screen is shown.
	if s.Data.OpenMenus.LoadingScreen {
		time.Sleep(time.Millisecond * 5)
	}

	for s.Priority != s.ExecutionPriority {
		if s.ExecutionPriority == PriorityStop {
			panic("Bot is stopped")
		}

		time.Sleep(time.Millisecond * 10)
	}
}
func (ctx *Context) WaitForGameToLoad() {
	for ctx.Data.OpenMenus.LoadingScreen {
		time.Sleep(100 * time.Millisecond)
		ctx.RefreshGameData()
	}
	// Add a small buffer to ensure everything is fully loaded
	time.Sleep(300 * time.Millisecond)
}
