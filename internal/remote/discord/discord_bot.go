package discord

import (
	"context"
	"fmt"
	"strings"

	koolo "github.com/hectorgimenez/koolo/internal"

	"github.com/bwmarrin/discordgo"
	"github.com/hectorgimenez/koolo/internal/event/stat"
)

type Bot struct {
	discordSession *discordgo.Session
	channelID      string
	supervisor     koolo.Supervisor
	companion      koolo.Companion
}

func NewBot(token, channelID string, supervisor koolo.Supervisor, companion koolo.Companion) (*Bot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	return &Bot{
		discordSession: dg,
		channelID:      channelID,
		supervisor:     supervisor,
		companion:      companion,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
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
	if strings.Contains(m.Content, "New game created.") {
		gameData := strings.SplitAfter(m.Content, "GameName: ")
		gameData = strings.Split(gameData[1], "|||")

		b.companion.JoinGame(gameData[0], gameData[1])
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	switch strings.ToLower(m.Content) {
	case "stats":
		b.publishStats()
	case "start":
		// TODO: Implement
	case "stop":
		b.supervisor.Stop()
	}
}

func (b *Bot) publishStats() {
	msg := "Run | Items | Deaths | Chickens | Merc Chickens | Errors | Healing Pots | Mana Pots | Reju Pots | Merc Pots \n"
	for run, st := range stat.Status.RunStats {
		msg += fmt.Sprintf(
			"%s | %d | %d | %d | %d | %d | %d | %d| %d | %d | %d \n",
			run,
			len(st.ItemsFound),
			st.Kills,
			st.Deaths,
			st.Chickens,
			st.MerChicken,
			st.Errors,
			st.HealingPotionsUsed,
			st.ManaPotionsUsed,
			st.RejuvPotionsUsed,
			st.MercHealingPotionsUsed,
		)
	}

	b.discordSession.ChannelMessageSend(b.channelID, msg)
}
