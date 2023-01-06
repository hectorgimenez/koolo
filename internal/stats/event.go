package stats

import (
	"context"
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"go.uber.org/zap"
	"image"
	"os"
	"time"
)

var Events = make(chan EventMsg, 10)

type EventMsg struct {
	Message string
	Image   image.Image
}

type EventListener struct {
	discordBot *DiscordBot
	logger     *zap.Logger
}

func EventWithScreenshot(message string) EventMsg {
	return EventMsg{
		Message: message,
		Image:   helper.Screenshot(),
	}
}

func NewEventListener(discordBot *DiscordBot, logger *zap.Logger) EventListener {
	return EventListener{
		discordBot: discordBot,
		logger:     logger,
	}
}

func (el EventListener) Listen(ctx context.Context) error {
	for {
		select {
		case e := <-Events:
			if _, err := os.Stat("screenshots"); os.IsNotExist(err) {
				err = os.MkdirAll("screenshots", 0700)
				if err != nil {
					el.logger.Error("error creating screenshots directory", zap.Error(err))
				}
			}

			fileName := fmt.Sprintf("screenshots/error-%s.png", time.Now().Format("2006-01-02 15_04_05"))
			err := robotgo.SavePng(e.Image, fileName)
			if err != nil {
				el.logger.Error("error saving screenshot", zap.Error(err))
			}

			if config.Config.Discord.Enabled {
				err = el.discordBot.SendMessageWithScreenshot(e.Message, e.Image)
				if err != nil {
					el.logger.Error("error sending message to discord", zap.Error(err))
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}
