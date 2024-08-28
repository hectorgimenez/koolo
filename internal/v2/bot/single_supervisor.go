package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/hectorgimenez/koolo/internal/game"
	ct "github.com/hectorgimenez/koolo/internal/v2/context"
	"github.com/hectorgimenez/koolo/internal/v2/health"
	"github.com/hectorgimenez/koolo/internal/v2/run"
	"github.com/hectorgimenez/koolo/internal/v2/utils"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
)

type SinglePlayerSupervisor struct {
	*baseSupervisor
}

func (s *SinglePlayerSupervisor) GetData() *game.Data {
	return s.bot.ctx.Data
}

func (s *SinglePlayerSupervisor) GetContext() *ct.Context {
	return s.bot.ctx
}

func NewSinglePlayerSupervisor(name string, bot *Bot, statsHandler *StatsHandler) (*SinglePlayerSupervisor, error) {
	bs, err := newBaseSupervisor(bot, name, statsHandler)
	if err != nil {
		return nil, err
	}

	return &SinglePlayerSupervisor{
		baseSupervisor: bs,
	}, nil
}

// Start will return error if it can not be started, otherwise will always return nil
func (s *SinglePlayerSupervisor) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel

	err := s.ensureProcessIsRunningAndPrepare()
	if err != nil {
		return fmt.Errorf("error preparing game: %w", err)
	}

	firstRun := true
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if firstRun {
				err = s.waitUntilCharacterSelectionScreen()
				if err != nil {
					return fmt.Errorf("error waiting for character selection screen: %w", err)
				}
			}
			if !s.bot.ctx.Manager.InGame() {

				// Create the game
				if err = s.HandleOutOfGameFlow(); err != nil {
					// Ignore loading screen errors
					if err.Error() == "loading screen" {
						continue
					}

					s.bot.ctx.Logger.Error(fmt.Sprintf("Error creating new game: %s", err.Error()))
					continue
				}
			}

			runs := run.BuildRuns(s.bot.ctx.CharacterCfg)
			gameStart := time.Now()
			if config.Characters[s.name].Game.RandomizeRuns {
				rand.Shuffle(len(runs), func(i, j int) { runs[i], runs[j] = runs[j], runs[i] })
			}
			if config.Koolo.Discord.EnableGameCreatedMessages {
				event.Send(event.GameCreated(event.Text(s.name, "New game created"), "", ""))
			} else {
				event.Send(event.GameCreated(event.Text(s.name, ""), "", ""))
			}
			s.bot.ctx.LastBuffAt = time.Time{}
			s.logGameStart(runs)

			// Refresh the Game data for a new game
			s.bot.ctx.GameReader.GetData(true)
			err = s.bot.Run(ctx, firstRun, runs)
			if err != nil {
				if errors.Is(context.Canceled, ctx.Err()) {
					continue
				}

				switch {
				case errors.Is(err, health.ErrChicken):
					if config.Koolo.Discord.EnableDiscordChickenMessages {
						event.Send(event.GameFinished(event.WithScreenshot(s.name, err.Error(), s.bot.ctx.GameReader.Screenshot()), event.FinishedChicken))
					} else {
						event.Send(event.GameFinished(event.Text(s.name, ""), event.FinishedChicken))
					}
					s.bot.ctx.Logger.Warn(err.Error(), slog.Float64("gameLength", time.Since(gameStart).Seconds()))
				case errors.Is(err, health.ErrMercChicken):
					if config.Koolo.Discord.EnableDiscordChickenMessages {
						event.Send(event.GameFinished(event.WithScreenshot(s.name, err.Error(), s.bot.ctx.GameReader.Screenshot()), event.FinishedMercChicken))
					} else {
						event.Send(event.GameFinished(event.Text(s.name, ""), event.FinishedMercChicken))
					}
					s.bot.ctx.Logger.Warn(err.Error(), slog.Float64("gameLength", time.Since(gameStart).Seconds()))
				case errors.Is(err, health.ErrDied):
					if config.Koolo.Discord.EnableDiscordChickenMessages {
						event.Send(event.GameFinished(event.WithScreenshot(s.name, err.Error(), s.bot.ctx.GameReader.Screenshot()), event.FinishedDied))
					} else {
						event.Send(event.GameFinished(event.Text(s.name, ""), event.FinishedDied))
					}
					s.bot.ctx.Logger.Warn(err.Error(), slog.Float64("gameLength", time.Since(gameStart).Seconds()))
				default:
					event.Send(event.GameFinished(event.WithScreenshot(s.name, err.Error(), s.bot.ctx.GameReader.Screenshot()), event.FinishedError))
					s.bot.ctx.Logger.Warn(
						fmt.Sprintf("Game finished with errors, reason: %s. Game total time: %0.2fs", err.Error(), time.Since(gameStart).Seconds()),
						slog.String("supervisor", s.name),
						slog.Uint64("mapSeed", uint64(s.bot.ctx.GameReader.CachedMapSeed)),
					)
				}
			}
			if exitErr := s.bot.ctx.Manager.ExitGame(); exitErr != nil {
				errMsg := fmt.Sprintf("Error exiting game %s", err.Error())
				event.Send(event.GameFinished(event.WithScreenshot(s.name, errMsg, s.bot.ctx.GameReader.Screenshot()), event.FinishedError))

				return errors.New(errMsg)
			}
			firstRun = false
		}
	}
}

// This function is responsible for handling all interactions with joining/creating games
func (s *SinglePlayerSupervisor) HandleOutOfGameFlow() error {

	// Catch-all
	if s.bot.ctx.CharacterCfg.AuthMethod != "None" && s.bot.ctx.GameReader.IsInCharacterSelectionScreen() && !s.bot.ctx.GameReader.IsOnline() {

		// Try and click the online tab to re-connect to bnet
		s.bot.ctx.HID.Click(game.LeftButton, 1090, 32) // click the online button

		// Wait a bit
		utils.Sleep(4000)

		// Check if we're online again, if not, kill the client
		if !s.bot.ctx.GameReader.IsOnline() {

			// Kill the client so the crash detector will restart it
			if err := s.KillClient(); err != nil {
				return err
			}

			return fmt.Errorf("we've lost connection to bnet or client glitched. The d2r process will be killed")
		}
	}

	// We're either in the in the Lobby or Character selection screen. Let's check
	if s.bot.ctx.GameReader.IsInCharacterSelectionScreen() {
		// TODO: Add Joining Games

		if s.bot.ctx.CharacterCfg.Game.CreateLobbyGames {
			retryCount := 0
			for !s.bot.ctx.GameReader.IsInLobby() {

				// Prevent an infinite loop
				if retryCount >= 5 && !s.bot.ctx.Data.IsInLobby {
					return fmt.Errorf("failed to enter bnet lobby after 5 retries")
				}

				// Try to enter bnet lobby
				s.bot.ctx.HID.Click(game.LeftButton, 744, 650)
				utils.Sleep(1000)
			}

			if _, err := s.bot.ctx.Manager.CreateOnlineGame(s.bot.ctx.CharacterCfg.Game.PublicGameCounter); err != nil {
				s.bot.ctx.CharacterCfg.Game.PublicGameCounter++
				return fmt.Errorf("failed to create an online game")

			} else {
				// We created the game successfully!
				s.bot.ctx.CharacterCfg.Game.PublicGameCounter++
				return nil
			}
		} else {
			// TODO: Add logic to check if we're on the online or offline tab and handle it accordingly.
			if !s.bot.ctx.GameReader.IsOnline() && s.bot.ctx.CharacterCfg.AuthMethod != "None" {

				// Try and click the online tab to re-connect to bnet
				s.bot.ctx.HID.Click(game.LeftButton, 1090, 32) // click the online button

				// Wait a bit
				utils.Sleep(4000)

				// Check again
				if !s.bot.ctx.GameReader.IsOnline() {
					// We failed to re-connect. Kill the client so it will get re-started automatically.
					if err := s.KillClient(); err != nil {
						return err
					}
					return fmt.Errorf("lost connection to bnet, killing client")
				}
			}

			// Create the game
			if err := s.bot.ctx.Manager.NewGame(); err != nil {
				return fmt.Errorf("failed to create game")
			}

			return nil
		}
	} else if s.bot.ctx.GameReader.IsInLobby() {
		// TODO: Add Joining Games

		// Check if we are suppose to create lobby games and enter lobby.
		if s.bot.ctx.CharacterCfg.Game.CreateLobbyGames {

			retryCount := 0
			for !s.bot.ctx.GameReader.IsInLobby() {

				// Prevent an infinite loop
				if retryCount >= 5 && !s.bot.ctx.GameReader.IsInLobby() {
					return fmt.Errorf("failed to enter bnet lobby after 5 retries")
				}

				// Try to enter bnet lobby
				s.bot.ctx.HID.Click(game.LeftButton, 744, 650)
				utils.Sleep(1000)
			}

			if _, err := s.bot.ctx.Manager.CreateOnlineGame(s.bot.ctx.CharacterCfg.Game.PublicGameCounter); err != nil {
				s.bot.ctx.CharacterCfg.Game.PublicGameCounter++
				return fmt.Errorf("failed to create an online game")

			} else {
				// We created the game successfully!
				s.bot.ctx.CharacterCfg.Game.PublicGameCounter++
				return nil
			}
		} else {
			// Press escape to exit the lobby
			s.bot.ctx.HID.PressKey(0x1B) // ESC - to avoid importing win here as well
			utils.Sleep(1000)

			for range 5 {
				if s.bot.ctx.Data.IsInCharSelectionScreen && s.bot.ctx.GameReader.IsOnline() {
					break
				}

				if s.bot.ctx.GameReader.IsInLobby() {
					// Mission failed
					s.bot.ctx.HID.PressKey(0x1B) // ESC - to avoid importing win here as well
					utils.Sleep(1000)
				}
			}

			if !s.bot.ctx.Data.IsInCharSelectionScreen {
				return fmt.Errorf("failed to leave lobby or an unknown case occurred")
			}

			// Create the game
			if err := s.bot.ctx.Manager.NewGame(); err != nil {
				return fmt.Errorf("failed to create game")
			}
		}
	} else if s.bot.ctx.Data.OpenMenus.LoadingScreen {
		// We're in a loading screen, wait a bit
		utils.Sleep(250)
		return fmt.Errorf("loading screen")
	} else {
		return fmt.Errorf("unhandled menu scenario")
	}

	// TODO: Maybe expand this with functionality to create new characters if the currently configured char isn't found? :)

	return nil
}
