package model

import (
	"strconv"
	"strings"
)

type Status string

var (
	ChannelStatusMain      Status = "main"
	ChannelStatusSecondary Status = "secondary"
)

type Channel struct {
	ID                int    `json:"id"`
	ChannelTelegramId int64  `json:"channel_telegram_id"`
	Name              string `json:"name"`
	URL               string `json:"url"`
	ChannelStatus     Status `json:"channel_status"`
}

func GetID(data string) int {
	parts := strings.Split(data, "_")
	if len(parts) > 3 {
		return 0
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0
	}

	return id
}
