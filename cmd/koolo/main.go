package main

import (
	"context"
	"errors"
	sloggger "github.com/hectorgimenez/koolo/cmd/koolo/log"
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/remote/discord"
	"github.com/hectorgimenez/koolo/internal/remote/telegram"
	"github.com/hectorgimenez/koolo/internal/server"
	"github.com/jchv/go-webview2"
	"golang.org/x/sync/errgroup"
	"log"
	"log/slog"
)

func main() {
	err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	logger, err := sloggger.NewLogger(config.Config.Debug.Log, config.Config.LogSaveDirectory)
	if err != nil {
		log.Fatalf("Error starting logger: %s", err.Error())
	}
	defer sloggger.FlushLog()

	ctx, cancel := context.WithCancel(context.Background())

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		w := webview2.NewWithOptions(webview2.WebViewOptions{
			Debug:     false,
			AutoFocus: true,
			WindowOptions: webview2.WindowOptions{
				Title:  "Koolo",
				Width:  1280,
				Height: 720,
				IconId: 1, // icon resource id
				Center: true,
			},
		})
		if w == nil {
			return errors.New("failed to load webview")
		}
		defer w.Destroy()
		w.SetSize(1280, 720, webview2.HintFixed)
		w.Navigate("http://localhost:8087")
		w.Run()

		cancel()
		return nil
	})

	additionalHandlers := make([]event.Handler, 0)
	// Discord Bot initialization
	if config.Config.Discord.Enabled {
		discordBot, err := discord.NewBot(config.Config.Discord.Token, config.Config.Discord.ChannelID)
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
	if config.Config.Telegram.Enabled {
		telegramBot, err := telegram.NewBot(config.Config.Telegram.Token, config.Config.Telegram.ChatID, logger)
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
