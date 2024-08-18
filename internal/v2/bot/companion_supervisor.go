package bot

// TODO companion sueprvisor
//
//import (
//	"context"
//	"errors"
//	"fmt"
//	"log/slog"
//	"time"
//
//	"github.com/hectorgimenez/koolo/internal/action"
//
//	"github.com/hectorgimenez/koolo/internal/container"
//
//	"github.com/hectorgimenez/koolo/internal/config"
//	"github.com/hectorgimenez/koolo/internal/event"
//	"github.com/hectorgimenez/koolo/internal/helper"
//	"github.com/hectorgimenez/koolo/internal/run"
//)
//
//type CompanionSupervisor struct {
//	*baseSupervisor
//}
//
//func NewCompanionSupervisor(name string, bot *Bot, runFactory *run.Factory, statsHandler *StatsHandler, c container.Container, pid uint32, hwnd uintptr) (*CompanionSupervisor, error) {
//	bs, err := newBaseSupervisor(bot, runFactory, name, statsHandler, c)
//	if err != nil {
//		return nil, err
//	}
//
//	return &CompanionSupervisor{
//		baseSupervisor: bs,
//	}, nil
//}
//
//// Start will return error if it can not be started, otherwise will always return nil
//func (s *CompanionSupervisor) Start() error {
//	ctx, cancel := context.WithCancel(context.Background())
//	s.cancelFn = cancel
//
//	err := s.ensureProcessIsRunningAndPrepare(ctx)
//	if err != nil {
//		return fmt.Errorf("error preparing game: %w", err)
//	}
//
//	gameCounter := 0
//	firstRun := true
//	err = s.waitUntilCharacterSelectionScreen()
//	if err != nil {
//		return fmt.Errorf("error waiting for character selection screen: %w", err)
//	}
//
//	for {
//		select {
//		case <-ctx.Done():
//			return nil
//		default:
//			if s.c.CharacterCfg.Companion.Leader {
//				time.Sleep(time.Second * 5)
//				gameName, err := s.c.Manager.CreateOnlineGame(gameCounter)
//				gameCounter++ // Sometimes game is created but error during join, so game name will be in use
//				if err != nil {
//					s.c.Logger.Error(fmt.Sprintf("Error creating new game: %s", err.Error()))
//					continue
//				}
//
//				if config.Koolo.Discord.EnableGameCreatedMessages {
//					event.Send(event.GameCreated(event.Text(s.name, "New game created: %s"), gameName, config.Characters[s.name].Companion.GamePassword))
//				} else {
//					event.Send(event.GameCreated(event.Text(s.name, ""), gameName, config.Characters[s.name].Companion.GamePassword))
//				}
//				err = s.startBot(ctx, s.runFactory.BuildRuns(), firstRun)
//				if err != nil {
//					return err
//				}
//				action.ResetBuffTime(s.Name())
//				firstRun = false
//			} else {
//				s.c.Logger.Debug("Waiting for new game to be created...")
//				evt := s.c.EventListener.WaitForEvent(ctx)
//				if gcEvent, ok := evt.(event.GameCreatedEvent); ok && gcEvent.Name != "" {
//					err = s.c.Manager.JoinOnlineGame(gcEvent.Name, gcEvent.Password)
//					if err != nil {
//						s.c.Logger.Error(err.Error())
//						continue
//					}
//
//					runs := s.runFactory.BuildRuns()
//					err = s.startBot(ctx, runs, firstRun)
//					firstRun = false
//					if err != nil {
//						return err
//					}
//					action.ResetBuffTime(s.Name())
//				}
//			}
//		}
//	}
//}
//
//func (s *CompanionSupervisor) startBot(ctx context.Context, runs []run.Run, firstRun bool) error {
//	gameStart := time.Now()
//	s.logGameStart(runs)
//	err := s.bot.Run(ctx, firstRun, runs)
//	if err != nil {
//		if errors.Is(context.Canceled, ctx.Err()) {
//			return nil
//		}
//		errorMsg := fmt.Sprintf("Game finished with errors, reason: %s. Game total time: %0.2fs", err.Error(), time.Since(gameStart).Seconds())
//		event.Send(event.GameFinished(event.WithScreenshot(s.name, errorMsg, s.c.Reader.Screenshot()), event.FinishedError))
//		s.c.Logger.Warn(errorMsg, slog.String("supervisor", s.name))
//	}
//	if exitErr := s.c.Manager.ExitGame(); exitErr != nil {
//		return fmt.Errorf("error exiting game: %s", exitErr)
//	}
//	firstRun = false
//
//	helper.Sleep(5000)
//
//	return nil
//}
