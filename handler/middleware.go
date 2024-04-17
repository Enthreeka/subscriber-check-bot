package handler

import (
	"context"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"subscriber-check-bot/repo"
)

func AdminMiddleware(service repo.UserRepo, next ViewFunc) ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		user, err := service.GetUserByID(ctx, update.FromChat().ID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return err
		}

		if user.Role == "admin" || user.Role == "superAdmin" {
			return next(ctx, bot, update)
		}

		return errors.New("user not admin")
	}
}
