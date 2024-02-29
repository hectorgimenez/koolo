package telegram

import (
	"bytes"
	"context"
	"image/jpeg"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hectorgimenez/koolo/internal/event"
)

func (b *Bot) Handle(_ context.Context, e event.Event) error {
	if e.Image() != nil {
		buf := new(bytes.Buffer)
		err := jpeg.Encode(buf, e.Image(), nil)
		if err != nil {
			return err
		}

		photo := tgbotapi.NewPhoto(b.chatID, tgbotapi.FileBytes{
			Name:  e.Message(),
			Bytes: buf.Bytes(),
		})
		photo.Caption = e.Message()

		_, err = b.bot.Send(photo)

		return err
	}

	_, err := b.bot.Send(tgbotapi.NewMessage(b.chatID, e.Message()))

	return err
}
