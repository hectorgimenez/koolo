package discord

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"image"
	"image/png"
)

type Bot struct {
	discordSession *discordgo.Session
	channelID      string
}

func NewBot(token string, channelID string) (*Bot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	return &Bot{
		discordSession: dg,
		channelID:      channelID,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	b.discordSession.AddHandler(onMessageCreated)
	b.discordSession.Identify.Intents = discordgo.IntentsGuildMessages
	err := b.discordSession.Open()
	if err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}

	// Wait until context is finished
	<-ctx.Done()

	return b.discordSession.Close()
}

func (b *Bot) SendMessageWithScreenshot(message string, screenshot image.Image) error {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, screenshot)
	if err != nil {
		return err
	}

	_, err = b.discordSession.ChannelMessageSendComplex(b.channelID, &discordgo.MessageSend{
		File:    &discordgo.File{Name: "Screenshot.png", ContentType: "image/png", Reader: buf},
		Content: message,
	})

	return err
}

func onMessageCreated(s *discordgo.Session, m *discordgo.MessageCreate) {
	// TODO: Handle commands
}
