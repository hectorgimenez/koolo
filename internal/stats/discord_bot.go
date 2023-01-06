package stats

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"image"
	"image/png"
	"os"
	"strings"
)

type DiscordBot struct {
	discordSession *discordgo.Session
	channelID      string
}

func NewDiscordBot(token string, channelID string) (*DiscordBot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	return &DiscordBot{
		discordSession: dg,
		channelID:      channelID,
	}, nil
}

func (b *DiscordBot) Start(ctx context.Context) error {
	b.discordSession.AddHandler(b.onMessageCreated)
	b.discordSession.Identify.Intents = discordgo.IntentsGuildMessages
	err := b.discordSession.Open()
	if err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}

	// Wait until context is finished
	<-ctx.Done()

	return b.discordSession.Close()
}

func (b *DiscordBot) SendMessageWithScreenshot(message string, screenshot image.Image) error {
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

func (b *DiscordBot) onMessageCreated(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	switch strings.ToLower(m.Content) {
	case "stats":
		b.publishStats()
	case "start":
		// TODO: Implement
	case "stop":
		os.Exit(0)
		// TODO: Implement correctly
	}
}

func (b *DiscordBot) publishStats() {
	msg := "Run | Items | Deaths | Chickens | Merc Chickens | Errors | Healing Pots | Mana Pots | Reju Pots | Merc Pots \n"
	for run, stat := range Status.RunStats {
		msg += fmt.Sprintf(
			"%s | %d | %d | %d | %d | %d | %d | %d| %d | %d | %d \n",
			run,
			len(stat.ItemsFound),
			stat.Kills,
			stat.Deaths,
			stat.Chickens,
			stat.MerChicken,
			stat.Errors,
			stat.HealingPotionsUsed,
			stat.ManaPotionsUsed,
			stat.RejuvPotionsUsed,
			stat.MercHealingPotionsUsed,
		)
	}

	b.discordSession.ChannelMessageSend(b.channelID, msg)
}
