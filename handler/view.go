package handler

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"subscriber-check-bot/pkg/logger"
	"subscriber-check-bot/pkg/store"
	"subscriber-check-bot/repo"
)

type ViewHandler struct {
	Log   *logger.Logger
	Store *store.Store

	ChRepo  repo.ChannelRepo
	MsgRepo repo.MessageRepo
}

func (v *ViewHandler) GetStart() ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		text := "Ну что, поехали?"

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Принять участие", "second_step"),
			))

		if _, err := bot.Send(msg); err != nil {
			v.Log.Error("failed to send message: %v", err)
			return err
		}

		return nil
	}
}

func (v *ViewHandler) AdminGetPanel() ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		text := "Список команд доступных администратору"

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Назначить главный канал", "set_main_channel"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Управление администраторами", "admin_role_setting"),
			),
		)

		if _, err := bot.Send(msg); err != nil {
			v.Log.Error("failed to send message: %v", err)
			return err
		}

		return nil
	}
}

func (v *ViewHandler) AdminCancelCommand() ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		v.Store.Delete(update.Message.Chat.ID)

		text := "Все команды отменены"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		if _, err := bot.Send(msg); err != nil {
			v.Log.Error("failed to send message: %v", err)
			return err
		}

		return nil
	}
}
