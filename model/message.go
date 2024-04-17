package model

type Message struct {
	ID         int     `json:"id"`
	Message    *string `json:"message"`
	FileID     *string `json:"file_id"`
	ButtonUrl  *string `json:"button_url"`
	ButtonText *string `json:"button_text"`
}
