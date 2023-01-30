package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Bot struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

func NewBot(token string, chatID int64) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		bot:    bot,
		chatID: chatID,
	}, nil
}
