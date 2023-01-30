package telegram

import (
	"context"
	"fmt"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hectorgimenez/koolo/internal/event/stat"
	"go.uber.org/zap"
)

type Bot struct {
	bot    *tgbotapi.BotAPI
	chatID int64
	logger *zap.Logger
}

func NewBot(token string, chatID int64, logger *zap.Logger) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		bot:    bot,
		chatID: chatID,
		logger: logger,
	}, nil
}

func (b *Bot) Start(_ context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil && update.Message.Chat.ID == b.chatID { // If we got a message
			switch strings.ToLower(update.Message.Text) {
			case "stats":
				if err := b.publishStats(); err != nil {
					b.logger.Error("error sending telegram message", zap.Error(err))
				}
			case "start":
				// TODO: Implement
			case "stop":
				os.Exit(0)
				// TODO: Implement correctly
			}
		}
	}

	return nil
}

func (b *Bot) publishStats() error {
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

	_, err := b.bot.Send(tgbotapi.NewMessage(b.chatID, msg))

	return err
}
