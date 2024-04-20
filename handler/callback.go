package handler

import (
	"context"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"subscriber-check-bot/model"
	"subscriber-check-bot/pkg/logger"
	"subscriber-check-bot/pkg/store"
	"subscriber-check-bot/repo"
)

type CallbackHandler struct {
	Log   *logger.Logger
	Store *store.Store

	ChRepo   repo.ChannelRepo
	MsgRepo  repo.MessageRepo
	UserRepo repo.UserRepo
}

func createChannelMarkup(channel []model.Channel, command string) (*tgbotapi.InlineKeyboardMarkup, error) {
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	buttonsPerRow := 1

	for i, el := range channel {
		var btn tgbotapi.InlineKeyboardButton

		if command == "user" {
			btn = tgbotapi.NewInlineKeyboardButtonURL(el.Name, el.URL)
		} else {
			btn = tgbotapi.NewInlineKeyboardButtonData(el.Name, fmt.Sprintf("channel_%s_%d", command, el.ID))
		}

		row = append(row, btn)

		if (i+1)%buttonsPerRow == 0 || i == len(channel)-1 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}

	}

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return &markup, nil
}

func (c *CallbackHandler) SecondStep() ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {

		channels, err := c.ChRepo.GetByStatus(ctx, model.ChannelStatusSecondary)
		if err != nil {
			c.Log.Error("SecondStep: ChRepo.GetByStatus: %v", err)
			HandleError(bot, update, "Временные неполадки на сервере, пытаемся исправить")
			return nil
		}

		if channels == nil {
			c.Log.Error("SecondStep: channels == nil")
			HandleError(bot, update, "Каналов не найдено")
			return nil
		}

		markup, err := createChannelMarkup(channels, "user")
		if err != nil {
			c.Log.Error("SecondStep: createChannelMarkup: %v", err)
			HandleError(bot, update, "Временные неполадки на сервере, пытаемся исправить")
			return nil
		}

		text := "Пожалуйста, подпишитесь на каналы"
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ReplyMarkup = markup

		if _, err := bot.Send(msg); err != nil {
			c.Log.Error("failed to send message: %v", err)
			return err
		}

		textSec := "После подписки на все каналы нажмите на кнопку - ГОТОВО"
		msgSec := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, textSec)
		msgSec.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("ГОТОВО", "ready")))

		if _, err := bot.Send(msgSec); err != nil {
			c.Log.Error("failed to send message: %v", err)
			return err
		}

		return nil
	}
}

func (c *CallbackHandler) Ready() ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		channels, err := c.ChRepo.GetByStatus(ctx, model.ChannelStatusSecondary)
		if err != nil {
			c.Log.Error("Ready: ChRepo.GetByStatus: %v", err)
			HandleError(bot, update, "Временные неполадки на сервере, пытаемся исправить")
			return nil
		}

		if channels == nil {
			c.Log.Error("Ready: channels == nil")
			HandleError(bot, update, "Каналов не найдено")
			return nil
		}

		isMember, err := isChatMember(bot, c.Log, channels, update.CallbackQuery.From.ID)
		if err != nil {
			c.Log.Error("Ready: isChatMember: %v", err)
			HandleError(bot, update, "Временные неполадки на сервере, пытаемся исправить")
			return nil
		}

		if !isMember {
			textThird := "Пожалуйста, подпишитесь на все представленные каналы"
			msgSec := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, textThird)

			if _, err := bot.Send(msgSec); err != nil {
				c.Log.Error("failed to send message: %v", err)
				return err
			}

			return nil
		} else {
			channel, err := c.ChRepo.GetByStatus(ctx, model.ChannelStatusMain)
			if err != nil {
				c.Log.Error("Ready: ChRepo.GetByStatus: %v", err)
				HandleError(bot, update, "Временные неполадки на сервере, пытаемся исправить")
				return nil
			}

			if channel == nil || len(channel) == 0 {
				c.Log.Error("Ready: channel == nil")
				HandleError(bot, update, "Каналов не найдено")
				return nil
			}

			createLink := tgbotapi.CreateChatInviteLinkConfig{
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: channel[0].ChannelTelegramId,
				},
				MemberLimit: 1,
			}

			response, err := bot.Request(createLink)
			if err != nil {
				c.Log.Error("update.MyChatMember.Chat: create link error: %v", err)
				return err
			}

			resultByte, err := response.Result.MarshalJSON()
			if err != nil {
				c.Log.Error("update.MyChatMember.Chat: response.Result.MarshalJSON: %v", err)
				return err
			}

			type InviteLink struct {
				InviteLink string `json:"invite_link"`
			}

			l := &InviteLink{}
			err = json.Unmarshal(resultByte, l)
			if err != nil {
				c.Log.Error("update.MyChatMember.Chat: json.Unmarshal: %v", err)
				return err
			}

			textThird := "Присоединяйся к секретному каналу:\n" + l.InviteLink
			msgSec := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, textThird)

			if _, err := bot.Send(msgSec); err != nil {
				c.Log.Error("failed to send message: %v", err)
				return err
			}
			return nil
		}
	}
}

func (c *CallbackHandler) AdminSetMainChannel() ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		channels, err := c.ChRepo.GetByStatus(ctx, model.ChannelStatusSecondary)
		if err != nil {
			c.Log.Error("Ready: ChRepo.GetByStatus: %v", err)
			HandleError(bot, update, "Временные неполадки на сервере, пытаемся исправить")
			return nil
		}

		if channels == nil {
			c.Log.Error("Ready: channels == nil")
			HandleError(bot, update, "Каналов не найдено")
			return nil
		}

		markup, err := createChannelMarkup(channels, "set")
		if err != nil {
			c.Log.Error("SecondStep: createChannelMarkup: %v", err)
			HandleError(bot, update, "Временные неполадки на сервере, пытаемся исправить")
			return nil
		}

		text := "Нажмите на канал, который будет являться 'главным' каналом"
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		msg.ReplyMarkup = markup
		if _, err := bot.Send(msg); err != nil {
			c.Log.Error("failed to send message: %v", err)
			return err
		}

		return nil
	}
}

func (c *CallbackHandler) AdminChooseMainChannel() ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		channelID := model.GetID(update.CallbackData())

		isExist, id, err := c.ChRepo.IsExistMainChannel(ctx)
		if err != nil {
			c.Log.Error("AdminChooseMainChannel: ChRepo.IsExistMainChannel: %v", err)
			HandleError(bot, update, "Временные неполадки на сервере")
			return nil
		}
		if isExist {
			if err := c.ChRepo.UpdateStatus(ctx, model.ChannelStatusSecondary, id); err != nil {
				c.Log.Error("AdminChooseMainChannel: ChRepo.UpdateStatus: %v", err)
				HandleError(bot, update, "Временные неполадки на сервере")
				return nil
			}
		}

		if err := c.ChRepo.UpdateStatus(ctx, model.ChannelStatusMain, channelID); err != nil {
			c.Log.Error("AdminChooseMainChannel: ChRepo.UpdateStatus: %v", err)
			HandleError(bot, update, "Временные неполадки на сервере")
			return nil
		}

		text := "Главный канал выбран"
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
		if _, err := bot.Send(msg); err != nil {
			c.Log.Error("failed to send message: %v", err)
			return err
		}

		return nil
	}
}

func (c *CallbackHandler) AdminRoleSetting() ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		text := "Управление администраторами"

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Назначить роль администратора", "admin_set_role"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Отозвать роль администратора", "admin_delete_role"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Посмотреть список администраторов", "admin_look_up"),
			),
		)

		if _, err := bot.Send(msg); err != nil {
			c.Log.Error("failed to send message: %v", err)
			return err
		}

		return nil
	}
}

func (c *CallbackHandler) AdminLookUp() ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		admin, err := c.UserRepo.GetAllAdmin(ctx)
		if err != nil {
			c.Log.Error("AdminLookUp: UserRepo.GetAllAdmin: %v", err)
			HandleError(bot, update, "Временные неполадки на сервере")
			return nil
		}

		adminByte, err := json.MarshalIndent(admin, "", "\t")
		if err != nil {
			c.Log.Error("AdminLookUp: json.MarshalIndent: %v", err)
			HandleError(bot, update, "Временные неполадки на сервере")
			return nil
		}

		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, string(adminByte))
		if _, err := bot.Send(msg); err != nil {
			c.Log.Error("failed to send message: %v", err)
			return err
		}

		return nil
	}
}

func (c *CallbackHandler) AdminDeleteRole() ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		text := "Напишите никнейм пользователя, у которого вы хотите отозвать права администратором.\nДля отмены команды" +
			"отправьте /cancel"

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)

		_, err := bot.Send(msg)
		if err != nil {
			c.Log.Error("failed to send message: %v", err)
			return err
		}

		c.Store.Set(store.AdminStore{
			TypeCommand: store.UserAdminDelete,
		}, update.CallbackQuery.Message.Chat.ID)

		return nil
	}
}

func (c *CallbackHandler) AdminSetRole() ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		text := "Напишите никнейм пользователя, которого вы хотите назначить администратором.\nДля отмены команды" +
			"отправьте /cancel"

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)

		_, err := bot.Send(msg)
		if err != nil {
			c.Log.Error("failed to send message: %v", err)
			return err
		}

		c.Store.Set(store.AdminStore{
			TypeCommand: store.UserAdminCreate,
		}, update.CallbackQuery.Message.Chat.ID)

		return nil
	}
}

func isChatMember(bot *tgbotapi.BotAPI, log *logger.Logger, channels []model.Channel, userID int64) (bool, error) {
	for _, el := range channels {
		channel := tgbotapi.ChatInfoConfig{
			tgbotapi.ChatConfig{
				ChatID: el.ChannelTelegramId,
			},
		}
		chat, err := bot.GetChat(channel)
		if err != nil {
			return false, err
		}

		cfg := tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: chat.ID,
				UserID: userID,
			},
		}

		chatMember, err := bot.GetChatMember(cfg)
		if err != nil {
			log.Error("error with chatID = %d:%v", chat.ID, err)
			return false, err
		}

		switch chatMember.Status {
		case "creator", "administrator", "member":
			continue
		default:
			return false, nil
		}

	}
	return true, nil
}
