package koolo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/reader"
	"github.com/hectorgimenez/koolo/internal/run"
	"go.uber.org/zap"
)

type Companion interface {
	JoinGame(gameName, password string)
}

type CompanionGameData struct {
	GameName string
	Password string
}

type CompanionSupervisor struct {
	baseSupervisor
	companionCh chan CompanionGameData
}

func (s *CompanionSupervisor) JoinGame(gameName, password string) {
	s.companionCh <- CompanionGameData{GameName: gameName, Password: password}
}

func NewCompanionSupervisor(logger *zap.Logger, bot *Bot, gr *reader.GameReader, gm *helper.GameManager) *CompanionSupervisor {
	return &CompanionSupervisor{
		baseSupervisor: baseSupervisor{
			logger: logger,
			bot:    bot,
			gr:     gr,
			gm:     gm,
		},
		companionCh: make(chan CompanionGameData),
	}
}

// Start will stay running during the application lifecycle, it will orchestrate all the required bot pieces
func (s *CompanionSupervisor) Start(ctx context.Context, factory *run.Factory) error {
	err := s.ensureProcessIsRunningAndPrepare()
	if err != nil {
		return fmt.Errorf("error preparing game: %w", err)
	}

	gameCounter := 0
	firstRun := true
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if config.Config.Companion.Leader {
				gameName, err := s.gm.CreateOnlineGame(gameCounter)
				gameCounter++ // Sometimes game is created but error during join, so game name will be in use
				if err != nil {
					s.logger.Error(fmt.Sprintf("Error creating new game: %s", err.Error()))
					continue
				}

				event.Events <- event.Text(fmt.Sprintf("New game created. GameName: " + gameName + "|||x"))

				err = s.startBot(ctx, factory.BuildRuns(), firstRun)
				firstRun = false
				if err != nil {
					return err
				}
			} else {
				for {
					s.logger.Debug("Waiting for new game to be created...")
					select {
					case gd := <-s.companionCh:
						err := s.gm.JoinOnlineGame(gd.GameName, gd.Password)
						if err != nil {
							s.logger.Error(err.Error())
							continue
						}

						runs := factory.BuildRuns()
						err = s.startBot(ctx, runs, firstRun)
						firstRun = false
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
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
		event.Events <- event.WithScreenshot(errorMsg)
		s.logger.Warn(errorMsg)
	}
	if exitErr := s.gm.ExitGame(); exitErr != nil {
		return fmt.Errorf("error exiting game: %s", exitErr)
	}
	firstRun = false

	s.updateGameStats()
	helper.Sleep(5000)

	return nil
}
