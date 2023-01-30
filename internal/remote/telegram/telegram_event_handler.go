package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hectorgimenez/koolo/internal/event"
)

func (b *Bot) Handle(_ context.Context, m event.Message) error {
	_, err := b.bot.Send(tgbotapi.NewMessage(b.chatID, m.Message))
}
