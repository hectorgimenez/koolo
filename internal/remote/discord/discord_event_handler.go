package discord

import (
	"bytes"
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgimenez/koolo/internal/event"
	"image/jpeg"
)

func (b *Bot) Handle(_ context.Context, m event.Message) error {
	if m.Image != nil {
		buf := new(bytes.Buffer)
		err := jpeg.Encode(buf, m.Image, &jpeg.Options{Quality: 80})
		if err != nil {
			return err
		}

		_, err = b.discordSession.ChannelMessageSendComplex(b.channelID, &discordgo.MessageSend{
			File:    &discordgo.File{Name: "Screenshot.jpeg", ContentType: "image/jpeg", Reader: buf},
			Content: m.Message,
		})

		return err
	}

	_, err := b.discordSession.ChannelMessageSend(b.channelID, m.Message)

	return err
}
