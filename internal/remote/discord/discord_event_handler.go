package discord

import (
	"bytes"
	"context"
	"image/jpeg"

	"github.com/bwmarrin/discordgo"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
)

func (b *Bot) Handle(_ context.Context, e event.Event) error {
	if b.shouldPublish(e) {
		buf := new(bytes.Buffer)
		err := jpeg.Encode(buf, e.Image(), &jpeg.Options{Quality: 80})
		if err != nil {
			return err
		}

		_, err = b.discordSession.ChannelMessageSendComplex(b.channelID, &discordgo.MessageSend{
			File:    &discordgo.File{Name: "Screenshot.jpeg", ContentType: "image/jpeg", Reader: buf},
			Content: e.Message(),
		})

		return err
	}

	_, err := b.discordSession.ChannelMessageSend(b.channelID, e.Message())

	return err
}

func (b *Bot) shouldPublish(e event.Event) bool {
	if e.Image() == nil {
		return false
	}

	switch evt := e.(type) {
	case event.GameFinishedEvent:
		if evt.Reason == event.FinishedChicken && !config.Koolo.Discord.EnableDiscordChickenMessages {
			return false
		}
		if evt.Reason == event.FinishedOK && !config.Koolo.Discord.EnableRunFinishMessages {
			return false
		}
		if evt.Reason == event.FinishedError && !config.Koolo.Discord.EnableGameCreatedMessages {
			return false
		}
	case event.GameCreatedEvent:
		if !config.Koolo.Discord.EnableGameCreatedMessages {
			return false
		}
	}

	return true
}
