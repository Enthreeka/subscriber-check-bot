package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func HandleError(bot *tgbotapi.BotAPI, update *tgbotapi.Update, messageError string) {
	msg := tgbotapi.NewMessage(update.FromChat().ID, messageError)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("failed to send message: %v\n", err)
	}
}
