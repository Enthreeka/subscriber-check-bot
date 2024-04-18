package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"os/signal"
	"subscriber-check-bot/config"
	"subscriber-check-bot/handler"
	"subscriber-check-bot/pkg/logger"
	"subscriber-check-bot/pkg/postgres"
	"subscriber-check-bot/pkg/store"
	"subscriber-check-bot/repo"
	"syscall"
)

func main() {
	log := logger.New()

	cfg, err := config.New()
	if err != nil {
		log.Fatal("Failed load config: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Fatal("failed to load token %v", err)
	}
	bot.Debug = false

	log.Info("Authorized on account %s", bot.Self.UserName)

	psql, err := postgres.New(context.Background(), 5, cfg.Postgres.URL)
	if err != nil {
		log.Fatal("failed to connect PostgreSQL: %v", err)
	}
	defer psql.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	chRepo := repo.NewChannelRepo(psql)
	msgRepo := repo.NewMessageRepo(psql)
	userRepo := repo.NewUserRepo(psql)

	tgStore := store.NewStore()

	viewHandler := handler.ViewHandler{Log: log, ChRepo: chRepo, MsgRepo: msgRepo, Store: tgStore}
	callbackHandler := handler.CallbackHandler{Log: log,
		Store:    tgStore,
		ChRepo:   chRepo,
		MsgRepo:  msgRepo,
		UserRepo: userRepo,
	}

	newBot := handler.NewBot(bot, log, chRepo, msgRepo, userRepo, tgStore)

	newBot.RegisterCommandView("start", viewHandler.GetStart())
	newBot.RegisterCommandView("secret", handler.AdminMiddleware(userRepo, viewHandler.AdminGetPanel()))
	newBot.RegisterCommandView("cancel", handler.AdminMiddleware(userRepo, viewHandler.AdminCancelCommand()))

	newBot.RegisterCommandCallback("second_step", callbackHandler.SecondStep())
	newBot.RegisterCommandCallback("ready", callbackHandler.Ready())

	newBot.RegisterCommandCallback("admin_role_setting", handler.AdminMiddleware(userRepo, callbackHandler.AdminRoleSetting()))
	newBot.RegisterCommandCallback("set_main_channel", handler.AdminMiddleware(userRepo, callbackHandler.AdminSetMainChannel()))
	newBot.RegisterCommandCallback("channel_set", handler.AdminMiddleware(userRepo, callbackHandler.AdminChooseMainChannel()))

	newBot.RegisterCommandCallback("admin_set_role", handler.AdminMiddleware(userRepo, callbackHandler.AdminSetRole()))
	newBot.RegisterCommandCallback("admin_delete_role", handler.AdminMiddleware(userRepo, callbackHandler.AdminDeleteRole()))
	newBot.RegisterCommandCallback("admin_look_up", handler.AdminMiddleware(userRepo, callbackHandler.AdminLookUp()))

	if err := newBot.Run(ctx); err != nil {
		log.Fatal("failed to run tgbot: %v", err)
	}
}
