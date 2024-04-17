package model

import "time"

type User struct {
	ID         int64     `json:"id,omitempty"`
	UsernameTg string    `json:"tg_username"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	Role       string    `json:"user_role"`
}
