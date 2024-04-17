package repo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"subscriber-check-bot/model"
	"subscriber-check-bot/pkg/postgres"
)

type MessageRepo interface {
	GetByID(ctx context.Context, id int) (*model.Message, error)

	Create(ctx context.Context, message *model.Message) error

	DeleteByID(ctx context.Context, id int) error

	UpdateTextByID(ctx context.Context, text *string, id int) error
	UpdateFileByID(ctx context.Context, fileID *string, fileType *string, id int) error
	UpdateButtonByID(ctx context.Context, buttonText *string, buttonURL *string, id int) error
}

type messageRepo struct {
	*postgres.Postgres
}

func NewMessageRepo(pg *postgres.Postgres) MessageRepo {
	return &messageRepo{
		pg,
	}
}

func (c *messageRepo) collectRow(row pgx.Row) (*model.Message, error) {
	var channel model.Message
	err := row.Scan(&channel.ID, &channel.Message, &channel.FileID, &channel.ButtonText, &channel.ButtonUrl)

	return &channel, err
}

func (c *messageRepo) collectRows(rows pgx.Rows) ([]model.Message, error) {
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Message, error) {
		message, err := c.collectRow(row)
		return *message, err
	})
}

func (m *messageRepo) GetByID(ctx context.Context, id int) (*model.Message, error) {
	q := `select id,message,file_id,button_url,button_text from message where id = $1`

	row := m.Pool.QueryRow(ctx, q, id)
	return m.collectRow(row)
}

func (m *messageRepo) Create(ctx context.Context, message *model.Message) error {
	q := `insert into message (message,file_id,button_url,button_text) values ($1,$2,$3,$4)`

	_, err := m.Pool.Exec(ctx, q, message.Message,
		message.FileID,
		message.ButtonUrl,
		message.ButtonText,
	)
	return err
}

func (m *messageRepo) DeleteByID(ctx context.Context, id int) error {
	q := `DELETE FROM message WHERE id = $1`

	_, err := m.Pool.Exec(ctx, q, id)
	return err
}

func (m *messageRepo) UpdateTextByID(ctx context.Context, text *string, id int) error {
	query := `update channel set notification_text = $1 where id = $2`

	_, err := m.Pool.Exec(ctx, query, text, id)
	return err
}

func (m *messageRepo) UpdateFileByID(ctx context.Context, fileID *string, fileType *string, id int) error {
	query := `update channek set file_id = $1, file_type = $2 where id = $3`

	_, err := m.Pool.Exec(ctx, query, fileID, fileType, id)
	return err
}

func (m *messageRepo) UpdateButtonByID(ctx context.Context, buttonText *string, buttonURL *string, id int) error {
	query := `update channel set button_url = $1, button_text = $2 where id = $3`

	_, err := m.Pool.Exec(ctx, query, buttonURL, buttonText, id)
	return err
}
