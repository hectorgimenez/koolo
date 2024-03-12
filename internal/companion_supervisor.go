package koolo

import (
	"context"
	"errors"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/container"
	"time"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/run"
)

type Companion interface {
	JoinGame(gameName, password string)
}

type CompanionGameData struct {
	GameName string
	Password string
}

type CompanionSupervisor struct {
	*baseSupervisor
	companionCh chan CompanionGameData
}

func (s *CompanionSupervisor) JoinGame(gameName, password string) {
	s.companionCh <- CompanionGameData{GameName: gameName, Password: password}
}

func NewCompanionSupervisor(name string, bot *Bot, runFactory *run.Factory, statsHandler *StatsHandler, listener *event.Listener, c container.Container) (*CompanionSupervisor, error) {
	bs, err := newBaseSupervisor(bot, runFactory, name, statsHandler, listener, c)
	if err != nil {
		return nil, err
	}

	return &CompanionSupervisor{
		baseSupervisor: bs,
		companionCh:    make(chan CompanionGameData),
	}, nil
}

// Start will return error if it can not be started, otherwise will always return nil
func (s *CompanionSupervisor) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel

	err := s.ensureProcessIsRunningAndPrepare(ctx)
	if err != nil {
		return fmt.Errorf("error preparing game: %w", err)
	}

	gameCounter := 0
	firstRun := true
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if config.Characters[s.name].Companion.Leader {
					gameName, err := s.c.Manager.CreateOnlineGame(gameCounter)
					gameCounter++ // Sometimes game is created but error during join, so game name will be in use
					if err != nil {
						s.c.Logger.Error(fmt.Sprintf("Error creating new game: %s", err.Error()))
						continue
					}

					s.c.EventChan <- event.GameCreated(event.Text("New game created: %s"), gameName, config.Characters[s.name].Companion.GamePassword)

					err = s.startBot(ctx, s.runFactory.BuildRuns(), firstRun)
					if err != nil {
						return
					}
					firstRun = false
				} else {
					for {
						s.c.Logger.Debug("Waiting for new game to be created...")
						select {
						case gd := <-s.companionCh:
							err := s.c.Manager.JoinOnlineGame(gd.GameName, gd.Password)
							if err != nil {
								s.c.Logger.Error(err.Error())
								continue
							}

							runs := s.runFactory.BuildRuns()
							err = s.startBot(ctx, runs, firstRun)
							firstRun = false
							if err != nil {
								return
							}
						}
					}
				}
			}
		}
	}()

	return nil
}

func (s *CompanionSupervisor) startBot(ctx context.Context, runs []run.Run, firstRun bool) error {
	gameStart := time.Now()
	s.logGameStart(runs)
	err := s.bot.Run(ctx, firstRun, runs)
	if err != nil {
		if errors.Is(context.Canceled, ctx.Err()) {
			return nil
		}
		errorMsg := fmt.Sprintf("Game finished with errors, reason: %s. Game total time: %0.2fs", err.Error(), time.Since(gameStart).Seconds())
		s.c.EventChan <- event.GameFinished(event.WithScreenshot(errorMsg, s.c.Reader.Screenshot()), event.FinishedError)
		s.c.Logger.Warn(errorMsg)
	}
	if exitErr := s.c.Manager.ExitGame(); exitErr != nil {
		return fmt.Errorf("error exiting game: %s", exitErr)
	}
	firstRun = false

	helper.Sleep(5000)

	return nil
}
