package main

import (
	"context"
	sloggger "github.com/hectorgimenez/koolo/cmd/koolo/log"
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/remote/discord"
	"github.com/hectorgimenez/koolo/internal/remote/telegram"
	"github.com/hectorgimenez/koolo/internal/server"
	"github.com/inkeliz/gowebview"
	"golang.org/x/sync/errgroup"
	"log"
	"log/slog"
	"os"
)

func main() {
	err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	logger, err := sloggger.NewLogger(config.Koolo.Debug.Log, config.Koolo.LogSaveDirectory)
	if err != nil {
		log.Fatalf("Error starting logger: %s", err.Error())
	}
	defer sloggger.FlushLog()

	ctx, cancel := context.WithCancel(context.Background())

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		w, err := gowebview.New(&gowebview.Config{URL: "http://localhost:8087", WindowConfig: &gowebview.WindowConfig{
			Title: "Koolo",
			Size: &gowebview.Point{
				X: 1280,
				Y: 720,
			},
		}})
		if err != nil {
			panic(err)
		}

		w.SetSize(&gowebview.Point{
			X: 1280,
			Y: 720,
		}, gowebview.HintFixed)

		defer w.Destroy()
		w.Run()

		cancel()
		os.Exit(0)
		return nil
	})

	additionalHandlers := make([]event.Handler, 0)
	// Discord Bot initialization
	if config.Koolo.Discord.Enabled {
		discordBot, err := discord.NewBot(config.Koolo.Discord.Token, config.Koolo.Discord.ChannelID)
		if err != nil {
			logger.Error("Discord could not been initialized", slog.Any("error", err))
			return
		}

		additionalHandlers = append(additionalHandlers, discordBot.Handle)
		g.Go(func() error {
			return discordBot.Start(ctx)
		})
	}

	// Telegram Bot initialization
	if config.Koolo.Telegram.Enabled {
		telegramBot, err := telegram.NewBot(config.Koolo.Telegram.Token, config.Koolo.Telegram.ChatID, logger)
		if err != nil {
			logger.Error("Telegram could not been initialized", slog.Any("error", err))
			return
		}

		additionalHandlers = append(additionalHandlers, telegramBot.Handle)
		g.Go(func() error {
			return telegramBot.Start(ctx)
		})
	}

	manager := koolo.NewSupervisorManager(logger, additionalHandlers)

	g.Go(func() error {
		srv := server.New(logger, manager)
		return srv.Listen(8087)
	})

	err = g.Wait()
	if err != nil {
		logger.Error("Error running Koolo", slog.Any("error", err))
		return
	}
}
