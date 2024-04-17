package repo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"subscriber-check-bot/model"
	"subscriber-check-bot/pkg/postgres"
)

type ChannelRepo interface {
	GetByID(ctx context.Context, id int) (*model.Channel, error)
	GetByName(ctx context.Context, name string) (*model.Channel, error)
	GetByStatus(ctx context.Context, status model.Status) ([]model.Channel, error)
	GetByChannelTelegramID(ctx context.Context, channelTelegramID int64) (*model.Channel, error)
	GetAll(ctx context.Context) ([]model.Channel, error)

	Create(ctx context.Context, channel *model.Channel) error

	DeleteByID(ctx context.Context, id int) error
	DeleteByName(ctx context.Context, name string) error

	UpdateStatus(ctx context.Context, status model.Status, id int) error

	IsExistMainChannel(ctx context.Context) (bool, int, error)
}

type channelRepo struct {
	*postgres.Postgres
}

func NewChannelRepo(pg *postgres.Postgres) ChannelRepo {
	return &channelRepo{
		pg,
	}
}

func (c *channelRepo) collectRow(row pgx.Row) (*model.Channel, error) {
	var channel model.Channel
	err := row.Scan(&channel.ID, &channel.ChannelTelegramId, &channel.Name, &channel.URL, &channel.ChannelStatus)
	//if errors.Is(err, pgx.ErrNoRows) {
	//	return nil, boterror.ErrNoRows
	//}
	//errCode := pgxError.ErrorCode(err)
	//if errCode == pgxError.ForeignKeyViolation {
	//	return nil, boterror.ErrForeignKeyViolation
	//}
	//if errCode == pgxError.UniqueViolation {
	//	return nil, boterror.ErrUniqueViolation
	//}
	return &channel, err
}

func (c *channelRepo) collectRows(rows pgx.Rows) ([]model.Channel, error) {
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Channel, error) {
		channel, err := c.collectRow(row)
		return *channel, err
	})
}

func (c *channelRepo) GetByID(ctx context.Context, id int) (*model.Channel, error) {
	q := `SELECT id,channel_telegram_id, name, url, channel_status FROM channel WHERE id = $1`

	row := c.Pool.QueryRow(ctx, q, id)
	return c.collectRow(row)
}

func (c *channelRepo) GetByName(ctx context.Context, name string) (*model.Channel, error) {
	q := `SELECT id,channel_telegram_id, name, url, channel_status FROM channel WHERE name = $1`

	row := c.Pool.QueryRow(ctx, q, name)
	return c.collectRow(row)
}

func (c *channelRepo) GetByStatus(ctx context.Context, status model.Status) ([]model.Channel, error) {
	q := `SELECT id,channel_telegram_id, name, url, channel_status FROM channel WHERE channel_status = $1`

	rows, err := c.Pool.Query(ctx, q, status)
	if err != nil {
		return nil, err
	}
	return c.collectRows(rows)
}

func (c *channelRepo) GetByChannelTelegramID(ctx context.Context, channelTelegramID int64) (*model.Channel, error) {
	q := `SELECT id,channel_telegram_id, name, url, channel_status FROM channel WHERE channel_telegram_id = $1`

	row := c.Pool.QueryRow(ctx, q, channelTelegramID)
	return c.collectRow(row)
}

func (c *channelRepo) GetAll(ctx context.Context) ([]model.Channel, error) {
	q := `SELECT id, name, url, channel_status FROM channel`

	rows, err := c.Pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	return c.collectRows(rows)
}

func (c *channelRepo) Create(ctx context.Context, channel *model.Channel) error {
	q := `INSERT INTO channel (channel_telegram_id,name, url, channel_status) VALUES ($1, $2, $3,$4)`

	_, err := c.Pool.Exec(ctx, q, channel.ChannelTelegramId,
		channel.Name,
		channel.URL,
		channel.ChannelStatus,
	)
	return err
}

func (c *channelRepo) DeleteByID(ctx context.Context, id int) error {
	q := `DELETE FROM channel WHERE id = $1`

	_, err := c.Pool.Exec(ctx, q, id)
	return err
}

func (c *channelRepo) DeleteByName(ctx context.Context, name string) error {
	q := `DELETE FROM channel WHERE name = $1`

	_, err := c.Pool.Exec(ctx, q, name)
	return err
}

func (c *channelRepo) UpdateStatus(ctx context.Context, status model.Status, id int) error {
	q := `update channel set channel_status = $1 where id = $2`

	_, err := c.Pool.Exec(ctx, q, status, id)

	return err
}

func (c *channelRepo) IsExistMainChannel(ctx context.Context) (bool, int, error) {
	q := `SELECT EXISTS (SELECT id FROM channel WHERE channel_status = 'main') AS exists_main, id FROM channel`

	var (
		isExist bool
		id      int
	)

	err := c.Pool.QueryRow(ctx, q).Scan(&isExist, &id)

	return isExist, id, err
}
