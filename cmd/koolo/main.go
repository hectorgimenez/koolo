package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	_ "net/http/pprof"
	"runtime/debug"

	sloggger "github.com/hectorgimenez/koolo/cmd/koolo/log"
	"github.com/hectorgimenez/koolo/internal/bot"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/remote/discord"
	"github.com/hectorgimenez/koolo/internal/remote/telegram"
	"github.com/hectorgimenez/koolo/internal/server"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/hectorgimenez/koolo/internal/utils/winproc"
	"github.com/inkeliz/gowebview"
	"golang.org/x/sync/errgroup"
)

func main() {
	err := config.Load()
	if err != nil {
		utils.ShowDialog("Error loading configuration", err.Error())
		log.Fatalf("Error loading configuration: %s", err.Error())
		return
	}

	logger, err := sloggger.NewLogger(config.Koolo.Debug.Log, config.Koolo.LogSaveDirectory, "")
	if err != nil {
		log.Fatalf("Error starting logger: %s", err.Error())
	}
	defer sloggger.FlushLog()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("fatal error detected, Koolo will close with the following error: %v\n Stacktrace: %s", r, debug.Stack())
			logger.Error(err.Error())
			sloggger.FlushLog()
			utils.ShowDialog("Koolo error :(", fmt.Sprintf("Koolo will close due to an expected error, please check the latest log file for more info!\n %s", err.Error()))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	winproc.SetProcessDpiAware.Call() // Set DPI awareness to be able to read the correct scale and show the window correctly

	eventListener := event.NewListener(logger)
	manager := bot.NewSupervisorManager(logger, eventListener)
	scheduler := bot.NewScheduler(manager, logger)
	go scheduler.Start()
	srv, err := server.New(logger, manager)
	if err != nil {
		log.Fatalf("Error starting local server: %s", err.Error())
	}

	g.Go(func() error {
		defer cancel()
		displayScale := config.GetCurrentDisplayScale()
		w, err := gowebview.New(&gowebview.Config{URL: "http://localhost:8087", WindowConfig: &gowebview.WindowConfig{
			Title: "Koolo",
			Size: &gowebview.Point{
				X: int64(1280 * displayScale),
				Y: int64(720 * displayScale),
			},
		}})
		if err != nil {
			w.Destroy()
			return fmt.Errorf("error creating webview: %w", err)
		}

		w.SetSize(&gowebview.Point{
			X: int64(1280 * displayScale),
			Y: int64(720 * displayScale),
		}, gowebview.HintFixed)

		defer w.Destroy()
		w.Run()

		return nil
	})

	// Discord Bot initialization
	if config.Koolo.Discord.Enabled {
		discordBot, err := discord.NewBot(config.Koolo.Discord.Token, config.Koolo.Discord.ChannelID, manager)
		if err != nil {
			logger.Error("Discord could not been initialized", slog.Any("error", err))
			return
		}

		eventListener.Register(discordBot.Handle)
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

		eventListener.Register(telegramBot.Handle)
		g.Go(func() error {
			return telegramBot.Start(ctx)
		})
	}

	g.Go(func() error {
		defer cancel()
		return srv.Listen(8087)
	})

	g.Go(func() error {
		defer cancel()
		return eventListener.Listen(ctx)
	})

	g.Go(func() error {
		<-ctx.Done()
		logger.Info("Koolo shutting down...")
		cancel()
		manager.StopAll()
		scheduler.Stop()
		err = srv.Stop()
		if err != nil {
			logger.Error("error stopping local server", slog.Any("error", err))
		}

		return err
	})

	err = g.Wait()
	if err != nil {
		cancel()
		logger.Error("Error running Koolo", slog.Any("error", err))
		return
	}

	sloggger.FlushLog()
}
