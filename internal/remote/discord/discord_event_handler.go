package discord

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"

	"github.com/bwmarrin/discordgo"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
)

func (b *Bot) Handle(_ context.Context, e event.Event) error {
	if b.shouldPublish(e) {

		switch e.(type) {
		case event.GameCreatedEvent, event.GameFinishedEvent, event.RunStartedEvent, event.RunFinishedEvent:
			_, err := b.discordSession.ChannelMessageSend(b.channelID, fmt.Sprintf("%s: %s", e.Supervisor(), e.Message()))
			return err
		default:
			break
		}

		buf := new(bytes.Buffer)
		err := jpeg.Encode(buf, e.Image(), &jpeg.Options{Quality: 80})
		if err != nil {
			return err
		}

		_, err = b.discordSession.ChannelMessageSendComplex(b.channelID, &discordgo.MessageSend{
			File:    &discordgo.File{Name: "Screenshot.jpeg", ContentType: "image/jpeg", Reader: buf},
			Content: fmt.Sprintf("%s: %s", e.Supervisor(), e.Message()),
		})

		return err
	}

	return nil
}

func (b *Bot) shouldPublish(e event.Event) bool {

	switch evt := e.(type) {
	case event.GameFinishedEvent:
		if evt.Reason == event.FinishedChicken || evt.Reason == event.FinishedMercChicken || evt.Reason == event.FinishedDied {
			return config.Koolo.Discord.EnableDiscordChickenMessages
		}
		if evt.Reason == event.FinishedOK {
			return false // supress game finished messages until we add proper option for it
		}
		return true
	case event.GameCreatedEvent:
		return config.Koolo.Discord.EnableGameCreatedMessages
	case event.RunStartedEvent:
		return config.Koolo.Discord.EnableNewRunMessages
	case event.RunFinishedEvent:
		return config.Koolo.Discord.EnableRunFinishMessages
	default:
		break
	}

	return e.Image() != nil
}
