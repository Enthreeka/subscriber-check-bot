package repo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"subscriber-check-bot/model"
	"subscriber-check-bot/pkg/postgres"
)

type UserRepo interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetAllUsers(ctx context.Context) ([]model.User, error)
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateRoleByUsername(ctx context.Context, role string, username string) error
	IsUserExistByUsernameTg(ctx context.Context, usernameTg string) (bool, error)
	GetAllAdmin(ctx context.Context) ([]model.User, error)
	IsUserExistByUserID(ctx context.Context, userID int64) (bool, error)
}

type userRepo struct {
	*postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) UserRepo {
	return &userRepo{
		pg,
	}
}

func (u *userRepo) collectRow(row pgx.Row) (*model.User, error) {
	var user model.User
	err := row.Scan(&user.ID, &user.UsernameTg, &user.CreatedAt, &user.Role)

	return &user, err
}

func (u *userRepo) collectRows(rows pgx.Rows) ([]model.User, error) {
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.User, error) {
		user, err := u.collectRow(row)
		return *user, err
	})
}

func (u *userRepo) CreateUser(ctx context.Context, user *model.User) error {
	query := `insert into "user" (id,tg_username,created_at,user_role) values ($1,$2,$3,$4)`

	_, err := u.Pool.Exec(ctx, query, user.ID, user.UsernameTg, user.CreatedAt, user.Role)
	return err
}

func (u *userRepo) GetAllUsers(ctx context.Context) ([]model.User, error) {
	query := `select * from "user"`

	rows, err := u.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return u.collectRows(rows)
}

func (u *userRepo) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	query := `select * from "user" where id = $1`

	row := u.Pool.QueryRow(ctx, query, id)
	return u.collectRow(row)
}

func (u *userRepo) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `select * from "user" where tg_username = $1`

	row := u.Pool.QueryRow(ctx, query, username)
	return u.collectRow(row)
}

func (u *userRepo) UpdateRoleByUsername(ctx context.Context, role string, username string) error {
	query := `update "user" set user_role = $1 where tg_username = $2`

	_, err := u.Pool.Exec(ctx, query, role, username)
	return err
}

func (u *userRepo) IsUserExistByUsernameTg(ctx context.Context, usernameTg string) (bool, error) {
	query := `select exists (select id from "user" where tg_username = $1)`
	var isExist bool

	err := u.Pool.QueryRow(ctx, query, usernameTg).Scan(&isExist)

	return isExist, err
}

func (u *userRepo) GetAllAdmin(ctx context.Context) ([]model.User, error) {
	query := `select * from "user" where user_role = 'admin' or user_role = 'superAdmin'`

	rows, err := u.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return u.collectRows(rows)
}

func (u *userRepo) IsUserExistByUserID(ctx context.Context, userID int64) (bool, error) {
	query := `select exists (select id from "user" where id = $1)`
	var isExist bool

	err := u.Pool.QueryRow(ctx, query, userID).Scan(&isExist)
	return isExist, err
}
