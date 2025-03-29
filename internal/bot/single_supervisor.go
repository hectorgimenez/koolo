package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/config"
	ct "github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/hectorgimenez/koolo/internal/utils"
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

	needToWait := true

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
			if firstRun && needToWait {
				err = s.waitUntilCharacterSelectionScreen()
				if err != nil {
					return fmt.Errorf("error waiting for character selection screen: %w", err)
				}
				needToWait = false
			}

			// By this point, we should be in the character selection screen.
			if !s.bot.ctx.Manager.InGame() {
				// Create the game
				if err = s.HandleMenuFlow(); err != nil {
					// Ignore loading screen errors or unhandled errors (for now) and try again
					if err.Error() == "loading screen" || err.Error() == "" {
						utils.Sleep(100)
						continue
					} else if err.Error() == "idle" {
						s.bot.ctx.Logger.Info("[Companion] Idling in character selection screen while waiting for Leader to create new game", slog.String("supervisor", s.name))
						utils.Sleep(100)
						continue
					}

					s.bot.ctx.Logger.Error(fmt.Sprintf("Error creating new game: %s", err.Error()))
					continue
				}
			}

			// Reset the companion game name and password to prevent re-joining the same game
			if s.bot.ctx.CharacterCfg.Companion.Enabled && !s.bot.ctx.CharacterCfg.Companion.Leader {
				s.bot.ctx.CharacterCfg.Companion.CompanionGameName = ""
				s.bot.ctx.CharacterCfg.Companion.CompanionGamePassword = ""
			}

			runs := run.BuildRuns(s.bot.ctx.CharacterCfg)
			gameStart := time.Now()
			if config.Characters[s.name].Game.RandomizeRuns {
				rand.Shuffle(len(runs), func(i, j int) { runs[i], runs[j] = runs[j], runs[i] })
			}

			event.Send(event.GameCreated(event.Text(s.name, "New game created"), s.bot.ctx.GameReader.LastGameName(), s.bot.ctx.GameReader.LastGamePass()))

			s.bot.ctx.LastBuffAt = time.Time{}
			s.logGameStart(runs)

			// Refresh game data to make sure we have the latest information
			s.bot.ctx.RefreshGameData()

			// If we're in companion mode, send the companion join game event
			if s.bot.ctx.CharacterCfg.Companion.Enabled && s.bot.ctx.CharacterCfg.Companion.Leader {
				event.Send(event.RequestCompanionJoinGame(event.Text(s.name, "New Game Started "+s.bot.ctx.Data.Game.LastGameName), s.bot.ctx.CharacterCfg.CharacterName, s.bot.ctx.Data.Game.LastGameName, s.bot.ctx.Data.Game.LastGamePassword))
			}

			// Perform keybindings check on the first run only
			if firstRun {
				missingKeybindings := s.bot.ctx.Char.CheckKeyBindings()
				if len(missingKeybindings) > 0 {
					var missingKeybindingsText = "Missing key binding for skill(s):"
					for _, v := range missingKeybindings {
						missingKeybindingsText += fmt.Sprintf("\n%s", skill.SkillNames[v])
					}
					missingKeybindingsText += "\nPlease bind the skills. Pausing bot..."

					utils.ShowDialog("Missing keybindings for "+s.bot.ctx.Name, missingKeybindingsText)
					s.TogglePause()
				}
			}

			err = s.bot.Run(ctx, firstRun, runs)
			firstRun = false

			var gameFinishReason event.FinishReason
			if err != nil {
				switch {
				case errors.Is(err, health.ErrChicken):
					gameFinishReason = event.FinishedChicken
				case errors.Is(err, health.ErrMercChicken):
					gameFinishReason = event.FinishedMercChicken
				case errors.Is(err, health.ErrDied):
					gameFinishReason = event.FinishedDied
				default:
					gameFinishReason = event.FinishedError
				}

				// Send the game finished event
				event.Send(event.GameFinished(event.WithScreenshot(s.name, err.Error(), s.bot.ctx.GameReader.Screenshot()), gameFinishReason))

				s.bot.ctx.Logger.Warn(
					fmt.Sprintf("Game finished with errors, reason: %s. Game total time: %0.2fs", err.Error(), time.Since(gameStart).Seconds()),
					slog.String("supervisor", s.name),
					slog.Uint64("mapSeed", uint64(s.bot.ctx.GameReader.MapSeed())),
				)
			} else {
				gameFinishReason = event.FinishedOK
				event.Send(event.GameFinished(event.Text(s.name, "Game finished successfully"), gameFinishReason))
			}

			// If we're in companion mode, send the companion stop join game event
			if s.bot.ctx.CharacterCfg.Companion.Enabled && s.bot.ctx.CharacterCfg.Companion.Leader {
				event.Send(event.ResetCompanionGameInfo(event.Text(s.name, "Game "+s.bot.ctx.Data.Game.LastGameName+" finished"), s.bot.ctx.CharacterCfg.CharacterName))
			}

			if exitErr := s.bot.ctx.Manager.ExitGame(); exitErr != nil {
				errMsg := fmt.Sprintf("Error exiting game %s", exitErr.Error())
				event.Send(event.GameFinished(event.WithScreenshot(s.name, errMsg, s.bot.ctx.GameReader.Screenshot()), event.FinishedError))
				return errors.New(errMsg)
			}
		}
	}
}

// This function is responsible for handling all interactions with joining/creating games
func (s *SinglePlayerSupervisor) HandleMenuFlow() error {

	s.bot.ctx.RefreshGameData()

	if s.bot.ctx.Data.OpenMenus.LoadingScreen {
		utils.Sleep(500)
		return fmt.Errorf("loading screen")
	}

	s.bot.ctx.Logger.Debug("[Menu Flow]: Starting menu flow ...")

	// Check if we're in character creation screen, and exit
	if s.bot.ctx.GameReader.IsInCharacterCreationScreen() {
		s.bot.ctx.Logger.Debug("[Menu Flow]: We're in character creation screen, exiting ...")

		// Click escape to exit character creation screen
		s.bot.ctx.HID.PressKey(0x1B)
		time.Sleep(2000)

		// Check if we're still in character creation screen
		if s.bot.ctx.GameReader.IsInCharacterCreationScreen() {
			return errors.New("[Menu Flow]: Failed to exit character creation screen")
		}
	}

	// Check if we're ingame for some reason but shouldn't?
	if s.bot.ctx.Manager.InGame() {
		s.bot.ctx.Logger.Debug("[Menu Flow]: We're still ingame, exiting ...")
		return s.bot.ctx.Manager.ExitGame()
	}

	// Check if there's any error popup
	isDismissableModalPresent, text := s.bot.ctx.GameReader.IsDismissableModalPresent()
	if isDismissableModalPresent {
		s.bot.ctx.Logger.Debug("[Menu Flow]: Detected dismissable modal with text: " + text)
		s.bot.ctx.Logger.Debug("[Menu Flow]: Dismissing it ...")

		// Click escape to dismiss the modal
		s.bot.ctx.HID.PressKey(0x1B)
		time.Sleep(1000)

		// Check again if the modal is still there
		isDismissableModalStillPresent, _ := s.bot.ctx.GameReader.IsDismissableModalPresent()
		if isDismissableModalStillPresent {
			return errors.New("[Menu Flow]: Failed to dismiss popup")
		}
	}

	// Check if we'll handle standard or companion mode
	if s.bot.ctx.CharacterCfg.Companion.Enabled && !s.bot.ctx.CharacterCfg.Companion.Leader {
		s.bot.ctx.Logger.Debug("[Menu Flow]: Companion mode detected, handling companion menu flow ...")
		return s.HandleCompanionMenuFlow()
	} else if s.bot.ctx.CharacterCfg.Companion.Enabled && s.bot.ctx.CharacterCfg.Companion.Leader {
		s.bot.ctx.Logger.Debug("[Menu Flow]: Companion Leader mode detected, using standard menu flow ...")
		return s.HandleStandardMenuFlow()
	} else {
		s.bot.ctx.Logger.Debug("[Menu Flow]: Standard mode detected, handling standard menu flow ...")
		return s.HandleStandardMenuFlow()
	}
}

func (s *SinglePlayerSupervisor) HandleStandardMenuFlow() error {

	atCharacterSelectionScreen := s.bot.ctx.GameReader.IsInCharacterSelectionScreen()

	if atCharacterSelectionScreen && s.bot.ctx.CharacterCfg.AuthMethod != "None" && !s.bot.ctx.CharacterCfg.Game.CreateLobbyGames {
		s.bot.ctx.Logger.Debug("[Menu Flow]: We're at the character selection screen, ensuring we're online ...")

		err := s.ensureOnline()
		if err != nil {
			return err
		}

		s.bot.ctx.Logger.Debug("[Menu Flow]: We're online, creating new game ...")

		// Create the game
		return s.bot.ctx.Manager.NewGame()

	} else if atCharacterSelectionScreen && s.bot.ctx.CharacterCfg.AuthMethod == "None" {

		s.bot.ctx.Logger.Debug("[Menu Flow]: Creating new game ...")
		return s.bot.ctx.Manager.NewGame()
	}

	atLobbyScreen := s.bot.ctx.GameReader.IsInLobby()

	if atLobbyScreen && s.bot.ctx.CharacterCfg.Game.CreateLobbyGames {
		s.bot.ctx.Logger.Debug("[Menu Flow]: We're at the lobby screen and we should create a lobby game ...")

		// Increment the game counter if its 0
		if s.bot.ctx.CharacterCfg.Game.PublicGameCounter == 0 {
			s.bot.ctx.CharacterCfg.Game.PublicGameCounter = 1
		}

		err := s.createLobbyGame()
		if err != nil {
			return err
		}

	} else if !atLobbyScreen && s.bot.ctx.CharacterCfg.Game.CreateLobbyGames {
		s.bot.ctx.Logger.Debug("[Menu Flow]: We're not at the lobby screen, trying to enter lobby ...")
		err := s.tryEnterLobby()
		if err != nil {
			return err
		}

		// Create the lobby game
		err = s.createLobbyGame()
		if err != nil {
			return err
		}
	} else if atLobbyScreen && !s.bot.ctx.CharacterCfg.Game.CreateLobbyGames {
		s.bot.ctx.Logger.Debug("[Menu Flow]: We're at the lobby screen, but we shouldn't be, going back to character selection screen ...")

		// Exit lobby by pressing esc
		s.bot.ctx.HID.PressKey(0x1B)
		time.Sleep(2000)

		// Check if we're still in lobby
		if s.bot.ctx.GameReader.IsInLobby() {
			return fmt.Errorf("[Menu Flow]: Failed to exit lobby")
		}

		// If we're at character selection screen, create a new game
		if s.bot.ctx.GameReader.IsInCharacterSelectionScreen() {
			return s.bot.ctx.Manager.NewGame()
		}
	}

	return fmt.Errorf("[Menu Flow]: Unhandled menu scenario")
}

func (s *SinglePlayerSupervisor) HandleCompanionMenuFlow() error {
	s.bot.ctx.Logger.Debug("[Menu Flow]: Trying to enter lobby ...")

	gameName := s.bot.ctx.CharacterCfg.Companion.CompanionGameName
	gamePassword := s.bot.ctx.CharacterCfg.Companion.CompanionGamePassword

	// If game name is blank, idle in menus
	if gameName == "" {
		// We don't have a game name, so we'll idle until we get one
		utils.Sleep(2000)
		return fmt.Errorf("idle")
	}

	if s.bot.ctx.GameReader.IsInCharacterSelectionScreen() {
		// Esnure we're online
		err := s.ensureOnline()
		if err != nil {
			return err
		}

		// Now we need to enter the lobby
		err = s.tryEnterLobby()
		if err != nil {
			return err
		}

		return s.bot.ctx.Manager.JoinOnlineGame(gameName, gamePassword)
	}

	if s.bot.ctx.GameReader.IsInLobby() {
		s.bot.ctx.Logger.Debug("[Menu Flow]: We're in lobby, joining game ...")
		return s.bot.ctx.Manager.JoinOnlineGame(gameName, gamePassword)
	}

	return fmt.Errorf("[Menu Flow]: Unhandled Companion menu scenario")
}

func (s *SinglePlayerSupervisor) tryEnterLobby() error {
	s.bot.ctx.Logger.Debug("[Menu Flow]: Trying to enter lobby ...")

	if s.bot.ctx.GameReader.IsInLobby() {
		s.bot.ctx.Logger.Debug("[Menu Flow]: We're already in lobby, exiting ...")
		return nil
	}

	retryCount := 0
	for !s.bot.ctx.GameReader.IsInLobby() {
		s.bot.ctx.Logger.Info("Entering lobby", slog.String("supervisor", s.name))
		// Prevent an infinite loop
		if retryCount >= 5 {
			return fmt.Errorf("[Menu Flow]: Failed to enter bnet lobby after 5 retries")
		}

		// Try to enter bnet lobby by clicking the "Play" button
		s.bot.ctx.HID.Click(game.LeftButton, 744, 650)
		utils.Sleep(1000)
		retryCount++
	}

	return nil
}

func (s *SinglePlayerSupervisor) createLobbyGame() error {
	s.bot.ctx.Logger.Debug("[Menu Flow]: Trying to create lobby game ...")

	// Create the online game
	_, err := s.bot.ctx.Manager.CreateLobbyGame(s.bot.ctx.CharacterCfg.Game.PublicGameCounter)
	if err != nil {
		s.bot.ctx.CharacterCfg.Game.PublicGameCounter++
		return fmt.Errorf("[Menu Flow]: Failed to create lobby game: %w", err)
	}

	// check if dismissable modal is present
	isDismissableModalPresent, text := s.bot.ctx.GameReader.IsDismissableModalPresent()
	if isDismissableModalPresent {
		s.bot.ctx.CharacterCfg.Game.PublicGameCounter++
		return fmt.Errorf("[Menu Flow]: Failed to create lobby game: %s", text)
	}

	s.bot.ctx.Logger.Debug("[Menu Flow]: Lobby game created successfully")

	// Game created successfully
	s.bot.ctx.CharacterCfg.Game.PublicGameCounter++
	return nil
}
