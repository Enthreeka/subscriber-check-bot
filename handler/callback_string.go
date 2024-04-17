package handler

import (
	"errors"
	"strings"
)

var (
	ErrNotFound = errors.New("not found in map")
)

func (b *Bot) CallbackStrings(callbackData string) (error, ViewFunc) {
	switch {
	case strings.HasPrefix(callbackData, "second_step"):
		callbackView, ok := b.callbackView["second_step"]
		if !ok {
			return ErrNotFound, nil
		}
		return nil, callbackView

	case strings.HasPrefix(callbackData, "ready"):
		callbackView, ok := b.callbackView["ready"]
		if !ok {
			return ErrNotFound, nil
		}
		return nil, callbackView

	case strings.HasPrefix(callbackData, "channel_get_"):
		callbackView, ok := b.callbackView["channel_get"]
		if !ok {
			return ErrNotFound, nil
		}
		return nil, callbackView

	case strings.HasPrefix(callbackData, "set_main_channel"):
		callbackView, ok := b.callbackView["set_main_channel"]
		if !ok {
			return ErrNotFound, nil
		}
		return nil, callbackView

	case strings.HasPrefix(callbackData, "channel_set_"):
		callbackView, ok := b.callbackView["channel_set"]
		if !ok {
			return ErrNotFound, nil
		}
		return nil, callbackView

	case strings.HasPrefix(callbackData, "admin_role_setting"):
		callbackView, ok := b.callbackView["admin_role_setting"]
		if !ok {
			return ErrNotFound, nil
		}
		return nil, callbackView

	case strings.HasPrefix(callbackData, "admin_set_role"):
		callbackView, ok := b.callbackView["admin_set_role"]
		if !ok {
			return ErrNotFound, nil
		}
		return nil, callbackView

	case strings.HasPrefix(callbackData, "admin_delete_role"):
		callbackView, ok := b.callbackView["admin_delete_role"]
		if !ok {
			return ErrNotFound, nil
		}
		return nil, callbackView

	case strings.HasPrefix(callbackData, "admin_look_up"):
		callbackView, ok := b.callbackView["admin_look_up"]
		if !ok {
			return ErrNotFound, nil
		}
		return nil, callbackView

	default:
		return nil, nil
	}
}
