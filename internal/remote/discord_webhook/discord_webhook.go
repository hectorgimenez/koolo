package discord_webhook

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gtuk/discordwebhook"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
)

type Webhook_Bot struct {
	discordWebhookURL string
}

type MessageContent struct {
	ItemName     string
	ItemQuality  string
	IsIdentified string
	IsEthereal   string
	ItemStats    string
	Rule         string
}

func NewWebhookBot(discordWebhookURL string) (*Webhook_Bot, error) {
	if discordWebhookURL == "" {
		return nil, fmt.Errorf("empty webhook URL")
	}

	return &Webhook_Bot{
		discordWebhookURL: discordWebhookURL,
	}, nil
}

func (b *Webhook_Bot) Start(ctx context.Context) error {
	// Wait until context is finished
	<-ctx.Done()

	return fmt.Errorf("no Error")
}

func (b *Webhook_Bot) Handle(_ context.Context, e event.Event) error {
	switch e.(type) {
	case event.ItemStashedEvent:
		evt, ok := e.(event.ItemStashedEvent)
		if !ok {
			return fmt.Errorf("discord Webhook: Failed to convert to ItemStashedEvent")
		}
		// If exclude filter matches, return
		match, _ := regexp.MatchString(config.Koolo.DiscordWebhook.Filter, evt.Item.Item.Desc().Name)
		if config.Koolo.DiscordWebhook.Filter != "" && match {
			return fmt.Errorf("discord Webhook: Item is filtered, not sending")
		}

		title := "Item Found!"
		color := "4289797"

		var statsStringBuf bytes.Buffer
		statsStringBuf.WriteString("\n")
		for _, element := range evt.Item.Item.Stats {
			statsStringBuf.WriteString(fmt.Sprintf("- %s: %d\n", stat.StringStats[element.ID], element.Value))
		}

		content := MessageContent{
			ItemName:     evt.Item.Item.Desc().Name,
			ItemQuality:  evt.Item.Item.Quality.ToString(),
			IsIdentified: strconv.FormatBool(evt.Item.Item.Identified),
			IsEthereal:   strconv.FormatBool(evt.Item.Item.Ethereal),
			ItemStats:    statsStringBuf.String(),
			Rule:         strings.ReplaceAll(evt.Item.Rule, "||", `\|\|`),
		}

		contentString := fmt.Sprintf("**Item:** %s\n**Quality:** %s\n**Identified:** %s\n**Ethereal:** %s\n**Rule:** %s\n**Stats:** %s", content.ItemName, content.ItemQuality, content.IsIdentified, content.IsEthereal, content.Rule, content.ItemStats)

		embed := discordwebhook.Embed{
			Title:       &title,
			Color:       &color,
			Description: &contentString,
		}
		embeds := []discordwebhook.Embed{
			embed,
		}

		charName := evt.Supervisor()
		message := discordwebhook.Message{
			Username: &charName,
			Embeds:   &embeds,
		}

		err := discordwebhook.SendMessage(b.discordWebhookURL, message)
		if err != nil {
			return fmt.Errorf("discord Webhook: Failed to send message")
		}
	}

	return fmt.Errorf("no Action taken")
}
