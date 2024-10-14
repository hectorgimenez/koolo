package bot

import (
	"context"
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/runtype"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action"
	botCtx "github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/run"
	"golang.org/x/sync/errgroup"
)

type Bot struct {
	ctx         *botCtx.Context
	runProgress *RunProgress
}

type RunProgress struct {
	VisitedAreas   map[area.ID]bool
	LastActionArea area.ID
	VisitedCoords  []data.Position
}

func NewBot(ctx *botCtx.Context) *Bot {
	return &Bot{
		ctx: ctx,
		runProgress: &RunProgress{
			VisitedAreas: make(map[area.ID]bool),
		},
	}
}
func (b *Bot) Run(ctx context.Context, firstRun bool, runs []runtype.Run) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	gameStartedAt := time.Now()
	b.ctx.SwitchPriority(botCtx.PriorityNormal)                                             // Restore priority to normal, in case it was stopped in previous game
	b.ctx.CurrentGame = &botCtx.CurrentGameHelper{ExpectedArea: b.ctx.Data.PlayerUnit.Area} // Reset current game helper structure

	err := b.ctx.GameReader.FetchMapData()
	if err != nil {
		return err
	}

	// Let's make sure we have updated game data also fully loaded before performing anything
	b.ctx.WaitForGameToLoad()
	// Switch to legacy mode if configured
	action.SwitchToLegacyMode()
	b.ctx.RefreshGameData()

	// This routine is in charge of refreshing the game data and handling cancellation, will work in parallel with any other execution
	g.Go(func() error {
		b.ctx.AttachRoutine(botCtx.PriorityBackground)
		ticker := time.NewTicker(10 * time.Millisecond)
		for {
			select {
			case <-ctx.Done():
				cancel()
				b.Stop()
				return nil
			case <-ticker.C:
				if b.ctx.ExecutionPriority == botCtx.PriorityPause {
					continue
				}
				b.ctx.RefreshGameData()
			}
		}
	})
	// This routine is in charge of handling the health/chicken of the bot, will work in parallel with any other execution
	g.Go(func() error {
		b.ctx.AttachRoutine(botCtx.PriorityBackground)
		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-ctx.Done():
				b.Stop()
				return nil
			case <-ticker.C:
				if b.ctx.ExecutionPriority == botCtx.PriorityPause {
					continue
				}
				err = b.ctx.HealthManager.HandleHealthAndMana()
				if err != nil {
					cancel()
					b.Stop()
					return err
				}
				if time.Since(gameStartedAt).Seconds() > float64(b.ctx.CharacterCfg.MaxGameLength) {
					cancel()
					b.Stop()
					return fmt.Errorf(
						"max game length reached, try to exit game: %0.2f",
						time.Since(gameStartedAt).Seconds(),
					)
				}
			}
		}
	})
	// High priority loop, this will interrupt (pause) low priority loop
	g.Go(func() error {
		defer func() {
			cancel()
			b.Stop()
			recover()
		}()

		b.ctx.AttachRoutine(botCtx.PriorityHigh)
		ticker := time.NewTicker(time.Millisecond * 100)
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if b.ctx.ExecutionPriority == botCtx.PriorityPause {
					continue
				}

				if b.ctx.CharacterCfg.ClassicMode && !b.ctx.Data.LegacyGraphics {
					action.SwitchToLegacyMode()
					b.ctx.RefreshGameData()
				}

				b.ctx.SwitchPriority(botCtx.PriorityHigh)

				// Area correction
				err := b.areaCorrection()
				if err != nil {
					b.ctx.Logger.Error("Area correction failed", "error", err)
				}
				// Perform item pickup if enabled
				if !b.ctx.DisableItemPickup {
					// Perform item pickup
					action.ItemPickup(30)
				}
				action.BuffIfRequired()

				_, healingPotsFound := b.ctx.Data.Inventory.Belt.GetFirstPotion(data.HealingPotion)
				_, manaPotsFound := b.ctx.Data.Inventory.Belt.GetFirstPotion(data.ManaPotion)
				// Check if we need to go back to town (no pots or merc died)
				if (b.ctx.CharacterCfg.BackToTown.NoHpPotions && !healingPotsFound ||
					b.ctx.CharacterCfg.BackToTown.EquipmentBroken && action.RepairRequired() ||
					b.ctx.CharacterCfg.BackToTown.NoMpPotions && !manaPotsFound ||
					b.ctx.CharacterCfg.BackToTown.MercDied && b.ctx.Data.MercHPPercent() <= 0 && b.ctx.CharacterCfg.Character.UseMerc) &&
					!b.ctx.Data.PlayerUnit.Area.IsTown() {
					action.InRunReturnTownRoutine()
				}
				b.ctx.SwitchPriority(botCtx.PriorityNormal)
			}
		}
	})

	// Low priority loop, this will keep executing main run scripts
	g.Go(func() error {
		defer func() {
			cancel()
			b.Stop()
			recover()
		}()

		b.ctx.AttachRoutine(botCtx.PriorityNormal)
		for k, r := range runs {
			event.Send(event.RunStarted(event.Text(b.ctx.Name, "Starting run"), r.Name()))
			b.ctx.CurrentGame.CurrentRun = r // Update the CurrentRun for area correction
			// Reset progress for each run
			b.ctx.CurrentGame.RunProgress = &botCtx.RunProgress{
				VisitedAreas:  make(map[area.ID]bool),
				VisitedCoords: make([]data.Position, 0),
			}
			err = action.PreRun(firstRun)
			if err != nil {
				return err
			}

			firstRun = false
			err = r.Run()
			if err != nil {
				return err
			}

			err = action.PostRun(k == len(runs)-1)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return g.Wait()
}

func (b *Bot) Stop() {
	b.ctx.SwitchPriority(botCtx.PriorityStop)
	b.ctx.Detach()
}

func (b *Bot) areaCorrection() error {
	// Skip correction if in town
	if b.ctx.Data.PlayerUnit.Area.IsTown() {
		return nil
	}
	// Skip correction if run isn't AreaAware
	areaAwareRun, isAreaAware := b.ctx.CurrentGame.CurrentRun.(run.AreaAwareRun)
	if !isAreaAware {
		return nil
	}

	currentArea := b.ctx.Data.PlayerUnit.Area
	expectedAreas := areaAwareRun.ExpectedAreas()
	visitedAreas := areaAwareRun.GetVisitedAreas()
	lastActionArea := areaAwareRun.GetLastActionArea()

	if !areaAwareRun.IsAreaPartOfRun(currentArea) {
		b.ctx.Logger.Info("Area mismatch detected", "current", currentArea.Area().Name, "lastAction", lastActionArea.Area().Name)

		if len(expectedAreas) == 0 {
			return fmt.Errorf("no expected areas defined for the current run")
		}

		// Check if any expected area is adjacent
		for _, adjacentLevel := range b.ctx.Data.AdjacentLevels {
			if areaAwareRun.IsAreaPartOfRun(adjacentLevel.Area) {
				err := action.MoveToArea(adjacentLevel.Area)
				if err == nil {
					visitedAreas[adjacentLevel.Area] = true
					lastActionArea = adjacentLevel.Area
					areaAwareRun.SetVisitedAreas(visitedAreas)
					areaAwareRun.SetLastActionArea(lastActionArea)
					return nil
				}
			}
		}

		// If no adjacent areas or move failed, return to town
		if err := action.ReturnTown(); err != nil {
			return fmt.Errorf("failed to return to town: %w", err)
		}

		// Find the last visited area in the expected sequence
		var lastVisitedIndex int
		for i, area := range expectedAreas {
			if visitedAreas[area] {
				lastVisitedIndex = i
			}
		}

		// From town, move sequentially through the expected areas up to the last visited area
		for i := 0; i <= lastVisitedIndex; i++ {
			targetArea := expectedAreas[i]

			var err error
			if i == 0 {
				if _, hasWaypoint := area.WPAddresses[targetArea]; hasWaypoint {
					err = action.WayPoint(targetArea)
				} else {
					err = action.MoveToArea(targetArea)
				}
			} else {
				err = action.MoveToArea(targetArea)
			}

			if err != nil {
				return fmt.Errorf("failed to move to area %s: %w", targetArea.Area().Name, err)
			}

			visitedAreas[targetArea] = true
			lastActionArea = targetArea
			areaAwareRun.SetVisitedAreas(visitedAreas)
			areaAwareRun.SetLastActionArea(lastActionArea)
		}
		// After moving to the correct area, move to the last visited coordinates
		if len(b.ctx.CurrentGame.RunProgress.VisitedCoords) > 0 {
			lastCoord := b.ctx.CurrentGame.RunProgress.VisitedCoords[len(b.ctx.CurrentGame.RunProgress.VisitedCoords)-1]
			err := action.MoveToCoords(lastCoord)
			if err != nil {
				return fmt.Errorf("failed to move to last known position: %w", err)
			}
			b.ctx.Logger.Info("Moved to last known position after area correction",
				"x", lastCoord.X, "y", lastCoord.Y)
		}

		b.ctx.Logger.Info("Area correction completed",
			"currentArea", b.ctx.Data.PlayerUnit.Area.Area().Name,
			"lastActionArea", areaAwareRun.GetLastActionArea().Area().Name)
	}

	return nil
}
