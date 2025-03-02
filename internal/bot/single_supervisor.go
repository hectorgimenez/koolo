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
				if err = s.HandleOutOfGameFlow(); err != nil {
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
			event.Send(event.GameCreated(event.Text(s.name, "ng:"+s.bot.ctx.Data.CharacterCfg.Companion.LeaderName+":")))
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
func (s *SinglePlayerSupervisor) HandleOutOfGameFlow() error {
	// Refresh the data
	s.bot.ctx.RefreshGameData()

	// First check if we're in a loading screen, which we can't do anything about
	if s.bot.ctx.Data.OpenMenus.LoadingScreen {
		utils.Sleep(250)
		return fmt.Errorf("loading screen")
	}

	// Check if we're in character creation screen, and handle it
	if s.bot.ctx.GameReader.IsInCharacterCreationScreen() {
		// We need to exit the character creation screen (press ESC)
		s.bot.ctx.HID.PressKey(0x1B) // ESC key
		utils.Sleep(1000)
		return fmt.Errorf("exiting character creation screen")
	}

	// Check if we need to be online but aren't
	if s.bot.ctx.GameReader.IsInCharacterSelectionScreen() && s.bot.ctx.CharacterCfg.AuthMethod != "None" && !s.bot.ctx.GameReader.IsOnline() {
		s.bot.ctx.Logger.Info("We're not online, trying to connect", slog.String("supervisor", s.name))
		err := s.EnsureOnlineTab()
		if err != nil {
			return err
		}
	}

	// Now handle based on whether we're in companion mode
	if s.bot.ctx.CharacterCfg.Companion.Enabled && !s.bot.ctx.CharacterCfg.Companion.Leader {
		return s.handleCompanionMode()
	} else {
		return s.handleNormalMode()
	}
}

// handleCompanionMode handles the flow for companion mode
func (s *SinglePlayerSupervisor) handleCompanionMode() error {

	// We are a follower/companion
	gameName := s.bot.ctx.CharacterCfg.Companion.CompanionGameName
	gamePassword := s.bot.ctx.CharacterCfg.Companion.CompanionGamePassword

	// If game name is blank, idle in character selection screen
	if gameName == "" {

		// Sleep for a bit
		utils.Sleep(2000)

		// We're in character selection screen and idling
		return fmt.Errorf("idle")
	}

	// We have a game name, so we need to join the leader's game
	if s.bot.ctx.GameReader.IsInCharacterSelectionScreen() {
		// Ensure we're on the online tab if using authentication
		err := s.EnsureOnlineTab()
		if err != nil {
			return err
		}

		// Need to go to lobby first to join a game
		err = s.enterLobby()
		if err != nil {
			return err
		}

		// Join the game
		return s.bot.ctx.Manager.JoinOnlineGame(gameName, gamePassword)
	} else if s.bot.ctx.GameReader.IsInLobby() {
		// We're in the lobby, join the game
		if err := s.bot.ctx.Manager.JoinOnlineGame(gameName, gamePassword); err != nil {
			return fmt.Errorf("failed to join leader's game: %w", err)
		}
		return nil
	}

	return fmt.Errorf("waiting for appropriate screen to join game")
}

// handleNormalMode handles the standard (non-companion) flow
func (s *SinglePlayerSupervisor) handleNormalMode() error {

	if s.bot.ctx.GameReader.IsInCharacterSelectionScreen() {
		// Ensure we're on the online tab if using authentication
		err := s.EnsureOnlineTab()
		if err != nil {
			return err
		}

		// If we need to create lobby games, go to lobby
		if s.bot.ctx.CharacterCfg.Game.CreateLobbyGames {
			err := s.enterLobby()
			if err != nil {
				return err
			}

			// Create Online Game
			return s.createLobbyGame()
		}

		// Otherwise, create a regular game from character selection
		if err := s.bot.ctx.Manager.NewGame(); err != nil {
			return fmt.Errorf("failed to create game: %w", err)
		}
		return nil

	} else if s.bot.ctx.GameReader.IsInLobby() {
		// If we should create lobby games, do so
		if s.bot.ctx.CharacterCfg.Game.CreateLobbyGames {
			return s.createLobbyGame()
		}

		// Otherwise, we need to go back to character selection
		return s.exitLobbyToCharacterSelection()
	}

	return fmt.Errorf("unknown game state")
}

// ensureOnlineTab makes sure we're on the online tab in character selection screen
func (s *SinglePlayerSupervisor) EnsureOnlineTab() error {
	if s.bot.ctx.CharacterCfg.AuthMethod != "None" && !s.bot.ctx.GameReader.IsOnline() {
		// Try and click the online tab to connect to bnet
		s.bot.ctx.HID.Click(game.LeftButton, 1090, 32)
		utils.Sleep(10000) // Wait for 10 seconds allowing time for the client to connect

		// Check if we're online again, if not, kill the client
		if !s.bot.ctx.GameReader.IsOnline() {
			if err := s.KillClient(); err != nil {
				return err
			}
			return fmt.Errorf("we've lost connection to bnet or client glitched. The d2r process will be killed")
		}
	}

	return nil
}

// enterLobby navigates to the lobby from character selection
func (s *SinglePlayerSupervisor) enterLobby() error {

	retryCount := 0
	for !s.bot.ctx.GameReader.IsInLobby() {
		s.bot.ctx.Logger.Info("Entering lobby", slog.String("supervisor", s.name))
		// Prevent an infinite loop
		if retryCount >= 5 {
			return fmt.Errorf("failed to enter bnet lobby after 5 retries")
		}

		// Try to enter bnet lobby by clicking the "Play" button
		s.bot.ctx.HID.Click(game.LeftButton, 744, 650)
		utils.Sleep(1000)
		retryCount++
	}

	return nil
}

// createLobbyGame creates a game from the lobby
func (s *SinglePlayerSupervisor) createLobbyGame() error {
	// Create the online game
	_, err := s.bot.ctx.Manager.CreateOnlineGame(s.bot.ctx.CharacterCfg.Game.PublicGameCounter)
	if err != nil {
		s.bot.ctx.CharacterCfg.Game.PublicGameCounter++
		return fmt.Errorf("failed to create an online game: %w", err)
	}

	// Game created successfully
	s.bot.ctx.CharacterCfg.Game.PublicGameCounter++
	return nil
}

// exitLobbyToCharacterSelection leaves the lobby and returns to character selection
func (s *SinglePlayerSupervisor) exitLobbyToCharacterSelection() error {
	// Press escape to exit the lobby
	s.bot.ctx.HID.PressKey(0x1B) // ESC key
	utils.Sleep(1000)

	// Retry a few times
	retryCount := 0
	for retryCount < 5 {
		if s.bot.ctx.GameReader.IsInCharacterSelectionScreen() &&
			(s.bot.ctx.GameReader.IsOnline() || s.bot.ctx.CharacterCfg.AuthMethod == "None") {
			break
		}

		if s.bot.ctx.GameReader.IsInLobby() {
			s.bot.ctx.HID.PressKey(0x1B) // ESC key
			utils.Sleep(1000)
		}

		retryCount++
	}

	if !s.bot.ctx.GameReader.IsInCharacterSelectionScreen() {
		return fmt.Errorf("failed to go back to character selection screen after multiple attempts")
	}

	// Now check if we need to be online but aren't
	if s.bot.ctx.CharacterCfg.AuthMethod != "None" && !s.bot.ctx.GameReader.IsOnline() {
		err := s.EnsureOnlineTab()
		if err != nil {
			return err
		}
	}

	// Create the game from character selection
	if err := s.bot.ctx.Manager.NewGame(); err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}

	return nil
}

func startGameCreationRoutine(s *SinglePlayerSupervisor) error {
	err := goToLobby(s)

	// TODO: this doesnt seem to actually kill the client. Gotta fix this.
	if err != nil {
		for err != nil {
			err = restartSupervisor(s)
		}
	}

	if _, err := s.bot.ctx.Manager.CreateOnlineGame(s.bot.ctx.CharacterCfg.Game.PublicGameCounter); err != nil {
		s.bot.ctx.CharacterCfg.Game.PublicGameCounter++
		return fmt.Errorf("failed to create an online game")

	} else {
		// We created the game successfully!
		s.bot.ctx.CharacterCfg.Game.PublicGameCounter++
		return nil
	}
}

var (
	LastGameJoined = ""
)

func startJoinerRoutine(s *SinglePlayerSupervisor) error {
	err := goToLobby(s)

	// TODO: this doesnt seem to actually kill the client. Gotta fix this.
	if err != nil {
		for err != nil {
			err = restartSupervisor(s)
		}
	}

	timeInLobby := time.Now()
	for LastGameJoined == config.LastGameName && !s.bot.ctx.Manager.InGame() && time.Since(timeInLobby) < 1*time.Minute {
		s.bot.ctx.Logger.Info("Waiting for game join")
		utils.Sleep(1000)
	}

	if s.bot.ctx.Manager.InGame() {
		LastGameJoined = s.bot.ctx.GameReader.LastGameName()
		return nil
	}

	for err := s.bot.ctx.Manager.JoinOnlineGame(config.LastGameName, config.LastGamePassword); err != nil && !s.bot.ctx.Manager.InGame(); err = s.bot.ctx.Manager.JoinOnlineGame(config.LastGameName, config.LastGamePassword) {
		// s.bot.ctx.HID.PressKey(0x1B)
		utils.Sleep(15000)
	}

	LastGameJoined = s.bot.ctx.GameReader.LastGameName()

	return nil
}

func goToLobby(s *SinglePlayerSupervisor) error {
	retryCount := 0
	for !s.bot.ctx.GameReader.IsInLobby() {

		// Prevent an infinite loop
		if retryCount >= 5 && !s.bot.ctx.Data.IsInLobby {
			return fmt.Errorf("failed to enter bnet lobby after 5 retries")
		}

		// Try to enter bnet lobby
		s.bot.ctx.HID.Click(game.LeftButton, 744, 650)
		utils.Sleep(1000)
		retryCount++
	}

	return nil
}

func restartSupervisor(s *SinglePlayerSupervisor) error {
	s.Stop()
	utils.Sleep(60000)
	return s.Start()
}
