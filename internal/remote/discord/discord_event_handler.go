package discord

import (
	"bytes"
	"context"
	"image/png"

	"github.com/bwmarrin/discordgo"
	"github.com/hectorgimenez/koolo/internal/event"
)

func (b *Bot) Handle(_ context.Context, m event.Message) error {
	if m.Image != nil {
		buf := new(bytes.Buffer)
		err := png.Encode(buf, m.Image)
		if err != nil {
			return err
		}

		_, err = b.discordSession.ChannelMessageSendComplex(b.channelID, &discordgo.MessageSend{
			File:    &discordgo.File{Name: "Screenshot.png", ContentType: "image/png", Reader: buf},
			Content: m.Message,
		})

		return err
	}

	_, err := b.discordSession.ChannelMessageSend(b.channelID, m.Message)

	return err
}
