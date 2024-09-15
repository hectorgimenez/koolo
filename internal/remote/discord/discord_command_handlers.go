package discord

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hectorgimenez/koolo/internal/bot"
)

func (b *Bot) supervisorExists(supervisor string) bool {
	supervisors := b.manager.AvailableSupervisors()
	return slices.Contains(supervisors, supervisor)
}

func (b *Bot) handleStartRequest(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Start supervisor(s) provided as in the stop command
	words := strings.Fields(m.Content)

	if len(words) > 1 {
		// Iterate through the supervisors specified
		for _, supervisor := range words[1:] {

			if !b.supervisorExists(supervisor) {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' not found.", supervisor))
				continue
			}

			// Attempt to start the specified supervisor
			b.manager.Start(supervisor, false)

			// Wait for the supervisor to start
			time.Sleep(1 * time.Second)

			// Send a confirmation message
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' has been started.", supervisor))
		}
	} else {
		// If no supervisors were specified, send a usage message
		s.ChannelMessageSend(m.ChannelID, "Usage: !start <supervisor1> [supervisor2] ...")
	}
}

func (b *Bot) handleStopRequest(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Split the message content into words
	words := strings.Fields(m.Content)

	// Check if there are any supervisors specified after "!stop"
	if len(words) > 1 {
		// Iterate through the supervisors specified
		for _, supervisor := range words[1:] {

			if !b.supervisorExists(supervisor) {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' not found.", supervisor))
				continue
			}

			// Check if the supervisor is running
			if b.manager.Status(supervisor).SupervisorStatus == bot.NotStarted || b.manager.Status(supervisor).SupervisorStatus == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' is not running.", supervisor))
				continue
			}

			// Attempt to stop the specified supervisor
			b.manager.Stop(supervisor)

			// Wait for the supervisor to stop
			time.Sleep(1 * time.Second)

			// Send a confirmation message
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' has been stopped.", supervisor))
		}
	} else {
		// If no supervisors were specified, send a usage message
		s.ChannelMessageSend(m.ChannelID, "Usage: !stop <supervisor1> [supervisor2] ...")
	}
}

func (b *Bot) handleStatusRequest(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Split the message content into words
	words := strings.Fields(m.Content)

	if len(words) > 1 {
		for _, supervisor := range words[1:] {
			if !b.supervisorExists(supervisor) {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' not found.", supervisor))
				continue
			}

			status := b.manager.Status(supervisor)
			if status.SupervisorStatus == bot.NotStarted || status.SupervisorStatus == "" {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' is offline.", supervisor))
				continue
			}

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' is %s", supervisor, status.SupervisorStatus))
		}
	} else {
		// If no supervisors were specified, send a usage message
		s.ChannelMessageSend(m.ChannelID, "Usage: !status <supervisor1> [supervisor2] ...")
	}
}

func (b *Bot) handleStatsRequest(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Split the message content into words
	words := strings.Fields(m.Content)

	// Check if there are any supervisors specified after "!stats"
	if len(words) > 1 {
		// Iterate through the supervisors specified
		for _, supervisor := range words[1:] {

			if !b.supervisorExists(supervisor) {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' not found.", supervisor))
				continue
			}

			// Fix for the status not being started
			supStatus := string(b.manager.Status(supervisor).SupervisorStatus)
			if supStatus == string(bot.NotStarted) || supStatus == "" {
				supStatus = "Offline"
			}
			// Create the embed
			embed := &discordgo.MessageEmbed{
				Title: fmt.Sprintf("Stats for %s", supervisor),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Status",
						Value:  supStatus,
						Inline: true,
					},
					{
						Name:   "Uptime",
						Value:  time.Since(b.manager.Status(supervisor).StartedAt).String(),
						Inline: true,
					},

					// Runs data

					{
						Name:   "Games",
						Value:  fmt.Sprintf("%d", b.manager.GetSupervisorStats(supervisor).TotalGames()),
						Inline: true,
					},
					{
						Name:   "Drops",
						Value:  fmt.Sprintf("%d", len(b.manager.GetSupervisorStats(supervisor).Drops)),
						Inline: true,
					},
					{
						Name:   "Deaths",
						Value:  fmt.Sprintf("%d", b.manager.GetSupervisorStats(supervisor).TotalDeaths()),
						Inline: true,
					},
					{
						Name:   "Chickens",
						Value:  fmt.Sprintf("%d", b.manager.GetSupervisorStats(supervisor).TotalChickens()),
						Inline: true,
					},
					{
						Name:   "Errors",
						Value:  fmt.Sprintf("%d", b.manager.GetSupervisorStats(supervisor).TotalErrors()),
						Inline: true,
					},
				},
			}

			// Send the embed to the channel
			s.ChannelMessageSendEmbed(m.ChannelID, embed)
		}
	} else {
		// If no supervisors were specified, send a usage message
		s.ChannelMessageSend(m.ChannelID, "Usage: !stats <supervisor1> [supervisor2] ...")
	}
}
