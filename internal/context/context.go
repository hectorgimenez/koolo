package context

import (
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
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
	ContextDebug      *Debug
	CurrentGame       *CurrentGameHelper
}

type Debug struct {
	LastAction string
	LastStep   string
}

type CurrentGameHelper struct {
	BlacklistedItems []data.Item
}

func NewContext(name string) *Status {
	ctx := &Context{
		Name:              name,
		Data:              &game.Data{},
		ExecutionPriority: PriorityNormal,
		ContextDebug:      &Debug{},
		CurrentGame:       &CurrentGameHelper{},
	}
	botContexts[getGoroutineID()] = &Status{Priority: PriorityNormal, Context: ctx}

	return botContexts[getGoroutineID()]
}

func Get() *Status {
	mu.Lock()
	defer mu.Unlock()
	return botContexts[getGoroutineID()]
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
		ctx.RefreshGameData()
		time.Sleep(100 * time.Millisecond)
	}
	// Add a small buffer to ensure everything is fully loaded
	time.Sleep(300 * time.Millisecond)
}
