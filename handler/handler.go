package handler

import (
	"context"
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"runtime/debug"
	"subscriber-check-bot/model"
	"subscriber-check-bot/pkg/logger"
	"subscriber-check-bot/pkg/store"
	"subscriber-check-bot/repo"
	"sync"
	"time"
)

const InternalServerError = "internal server error"

type ViewFunc func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error

type Bot struct {
	bot   *tgbotapi.BotAPI
	log   *logger.Logger
	store *store.Store

	chRepo   repo.ChannelRepo
	msgRepo  repo.MessageRepo
	userRepo repo.UserRepo

	cmdView      map[string]ViewFunc
	callbackView map[string]ViewFunc

	mu      sync.RWMutex
	isDebug bool
}

func NewBot(bot *tgbotapi.BotAPI,
	log *logger.Logger,
	chRepo repo.ChannelRepo,
	msgRepo repo.MessageRepo,
	userRepo repo.UserRepo,
	store *store.Store,
) *Bot {
	return &Bot{
		bot:      bot,
		log:      log,
		chRepo:   chRepo,
		msgRepo:  msgRepo,
		userRepo: userRepo,
		store:    store,
	}
}

func (b *Bot) RegisterCommandView(cmd string, view ViewFunc) {
	if b.cmdView == nil {
		b.cmdView = make(map[string]ViewFunc)
	}

	b.cmdView[cmd] = view
}

func (b *Bot) RegisterCommandCallback(callback string, view ViewFunc) {
	if b.callbackView == nil {
		b.callbackView = make(map[string]ViewFunc)
	}

	b.callbackView[callback] = view
}

func (b *Bot) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)
	for {
		select {
		case update := <-updates:
			updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

			b.isDebug = false
			b.jsonDebug(update.MyChatMember)

			b.handlerUpdate(updateCtx, &update)

			cancel()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *Bot) jsonDebug(update any) {
	if b.isDebug {
		updateByte, err := json.MarshalIndent(update, "", " ")
		if err != nil {
			b.log.Error("%v", err)
		}
		b.log.Info("%s", updateByte)
	}
}

func (b *Bot) handlerUpdate(ctx context.Context, update *tgbotapi.Update) {
	defer func() {
		if p := recover(); p != nil {
			b.log.Error("panic recovered: %v, %s", p, string(debug.Stack()))
		}
	}()

	// if write message
	if update.Message != nil {
		b.log.Info("[%s] %s", update.Message.From.UserName, update.Message.Text)

		isUserExist, err := b.userRepo.IsUserExistByUserID(ctx, update.Message.From.ID)
		if err != nil {
			b.log.Error("userRepo.IsUserExistByUserID: %v", err)
			HandleError(b.bot, update, InternalServerError)
			return
		}
		if !isUserExist {
			if err := b.userRepo.CreateUser(ctx, &model.User{
				ID:         update.Message.From.ID,
				UsernameTg: update.Message.From.UserName,
				CreatedAt:  time.Now(),
				Role:       "user",
			}); err != nil {
				b.log.Error("userService.CreateUser: failed to create user: %v", err)
				return
			}
		}

		isStoreExist := b.isStoreExist(ctx, update)
		if isStoreExist {
			return
		}

		var view ViewFunc

		cmd := update.Message.Command()

		cmdView, ok := b.cmdView[cmd]
		if !ok {
			return
		}

		view = cmdView

		if err := view(ctx, b.bot, update); err != nil {
			b.log.Error("failed to handle update: %v", err)
			HandleError(b.bot, update, InternalServerError)
			return
		}
		//  if press button
	} else if update.CallbackQuery != nil {
		b.log.Info("[%s] %s", update.CallbackQuery.From.UserName, update.CallbackData())

		var callback ViewFunc

		err, callbackView := b.CallbackStrings(update.CallbackData())
		if err != nil {
			b.log.Error("%v", err)
			return
		}

		callback = callbackView

		if err := callback(ctx, b.bot, update); err != nil {
			b.log.Error("failed to handle update: %v", err)
			HandleError(b.bot, update, InternalServerError)
			return
		}
		// if request on join chat
	} else if update.ChatJoinRequest != nil {
		b.log.Info("[%s] %s", update.ChatJoinRequest.From.UserName, update.ChatJoinRequest.InviteLink.InviteLink)

		// if bot update/delete from channel
	} else if update.MyChatMember != nil {

		if update.MyChatMember.Chat.IsChannel() {
			b.log.Info("[%s] %s", update.MyChatMember.From.UserName, update.MyChatMember.NewChatMember.Status)

			if update.MyChatMember.NewChatMember.Status == "administrator" {

				var link string
				if update.MyChatMember.Chat.InviteLink == "" {
					createLink := tgbotapi.CreateChatInviteLinkConfig{
						ChatConfig: tgbotapi.ChatConfig{
							ChatID: update.MyChatMember.Chat.ID,
						},
						CreatesJoinRequest: true,
					}

					response, err := b.bot.Request(createLink)
					if err != nil {
						b.log.Error("update.MyChatMember.Chat: create link error: %v", err)
						return
					}

					resultByte, err := response.Result.MarshalJSON()
					if err != nil {
						b.log.Error("update.MyChatMember.Chat: response.Result.MarshalJSON: %v", err)
						return
					}

					type InviteLink struct {
						InviteLink string `json:"invite_link"`
					}

					l := &InviteLink{}
					err = json.Unmarshal(resultByte, l)
					if err != nil {
						b.log.Error("update.MyChatMember.Chat: json.Unmarshal: %v", err)
						return
					}
					link = l.InviteLink
				} else {
					link = update.MyChatMember.Chat.InviteLink
				}

				if err := b.chRepo.Create(ctx, &model.Channel{
					Name:              update.MyChatMember.Chat.Title,
					URL:               link,
					ChannelStatus:     model.ChannelStatusSecondary,
					ChannelTelegramId: update.MyChatMember.Chat.ID,
				}); err != nil {
					b.log.Error("update.MyChatMember.Chat: chRepo.Create: %v", err)
					return
				}
			}

			if update.MyChatMember.NewChatMember.Status == "kicked" || update.MyChatMember.NewChatMember.Status == "left" {
				if err := b.chRepo.DeleteByName(ctx, update.MyChatMember.Chat.Title); err != nil {
					b.log.Error("update.MyChatMember.Chat: chRepo.DeleteByName: %v", err)
					return
				}
			}
		}

	}
}

func (b *Bot) isStoreExist(ctx context.Context, update *tgbotapi.Update) bool {
	userID := update.Message.Chat.ID
	data, exist := b.store.Read(userID)
	if !exist {
		return false
	}

	switch s := data.(type) {
	case store.AdminStore:
		defer b.store.Delete(userID)

		if s.TypeCommand == store.UserAdminCreate {
			if err := b.userRepo.UpdateRoleByUsername(ctx, "admin", update.Message.Text); err != nil {
				b.log.Error("isStoreExist:userRepo.UpdateRoleByUsername: %v", err)
				HandleError(b.bot, update, InternalServerError)
				return true
			}
			return true
		}
		if s.TypeCommand == store.UserAdminDelete {
			if err := b.userRepo.UpdateRoleByUsername(ctx, "user", update.Message.Text); err != nil {
				b.log.Error("isStoreExist:userRepo.UpdateRoleByUsername: %v", err)
				HandleError(b.bot, update, InternalServerError)
				return true
			}
			return true
		}

		b.log.Error("isStoreExist: undefind type command")
		return true
	default:
		b.log.Error("isStoreExist: type switching error")
		return false
	}
}
