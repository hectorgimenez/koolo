package discord

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/hectorgimenez/koolo/internal/bot"
	"github.com/hectorgimenez/koolo/internal/config"
)

type Bot struct {
	discordSession *discordgo.Session
	channelID      string
	manager        *bot.SupervisorManager
}

func NewBot(token, channelID string, manager *bot.SupervisorManager) (*Bot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	botInstance := &Bot{
		discordSession: dg,
		channelID:      channelID,
		manager:        manager,
	}

	return botInstance, nil
}

func (b *Bot) Start(ctx context.Context) error {
	//b.discordSession.Debug = true
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

func (b *Bot) onMessageCreated(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Check if the message follows the format "New game created:gameName:gamePassword"
	if m.Content == "" {
		chanMsgs, err := s.ChannelMessages(m.ChannelID, 1, "", "", m.ID)
		if err != nil {
			return
		}
		m.Content = chanMsgs[0].Content
		m.Attachments = chanMsgs[0].Attachments
	}

	if (m.Author.ID == s.State.User.ID || slices.Contains(config.Koolo.Discord.BotAdmins, m.Author.ID)) && strings.Contains(m.Content, "ng:") {
		parts := strings.Split(m.Message.Content, ":")
		if len(parts) == 3 {
			config.LastGameName = parts[1]
			config.LastGamePassword = parts[2]
		}
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if the message is from a bot admin
	if !slices.Contains(config.Koolo.Discord.BotAdmins, m.Author.ID) {
		return
	}

	prefix := strings.Split(m.Content, " ")[0]
	switch prefix {
	case "!start":
		b.handleStartRequest(s, m)
	case "!stop":
		b.handleStopRequest(s, m)
	case "!stats":
		b.handleStatsRequest(s, m)
	case "!status":
		b.handleStatusRequest(s, m)
	}

}
