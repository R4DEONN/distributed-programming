package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"userservice/pkg/user/domain/model"
)

func NewUserRepository(ctx context.Context, client mysql.ClientContext) model.UserRepository {
	return &userRepository{
		ctx:    ctx,
		client: client,
	}
}

type userRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (u *userRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (u *userRepository) Store(user model.User) error {
	_, err := u.client.ExecContext(u.ctx,
		`
	INSERT INTO user (user_id, status, login, email, telegram, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		status=VALUES(status),
	    login=VALUES(login),
	    email=VALUES(email),
	    telegram=VALUES(telegram),
	    updated_at=VALUES(updated_at),
	    deleted_at=VALUES(deleted_at)
	`,
		user.UserID,
		user.Status,
		user.Login,
		toSQLNull(user.Email),
		toSQLNull(user.Telegram),
		user.CreatedAt,
		user.UpdatedAt,
		toSQLNull(user.DeletedAt),
	)
	return errors.WithStack(err)
}

func (u *userRepository) Find(spec model.FindSpec) (*model.User, error) {
	user := struct {
		UserID    uuid.UUID           `db:"user_id"`
		Status    int                 `db:"status"`
		Login     string              `db:"login"`
		Email     sql.Null[string]    `db:"email"`
		Telegram  sql.Null[string]    `db:"telegram"`
		CreatedAt time.Time           `db:"created_at"`
		UpdatedAt time.Time           `db:"updated_at"`
		DeletedAt sql.Null[time.Time] `db:"deleted_at"`
	}{}
	query, args := u.buildSpecArgs(spec)

	err := u.client.GetContext(
		u.ctx,
		&user,
		`SELECT user_id, status, login, email, telegram, created_at, updated_at, deleted_at FROM user WHERE `+query,
		args...,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrUserNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &model.User{
		UserID:    user.UserID,
		Status:    model.UserStatus(user.Status),
		Login:     user.Login,
		Email:     fromSQLNull(user.Email),
		Telegram:  fromSQLNull(user.Telegram),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: fromSQLNull(user.DeletedAt),
	}, nil
}

func (u *userRepository) HardDelete(userID uuid.UUID) error {
	_, err := u.client.ExecContext(u.ctx, `DELETE FROM user WHERE user_id = ?`, userID)
	return errors.WithStack(err)
}

func (u *userRepository) buildSpecArgs(spec model.FindSpec) (query string, args []interface{}) {
	var parts []string
	if spec.UserID != nil {
		parts = append(parts, "user_id = ?")
		args = append(args, *spec.UserID)
	}
	if spec.Login != nil {
		parts = append(parts, "login = ?")
		args = append(args, *spec.Login)
	}
	if spec.Email != nil {
		parts = append(parts, "email = ?")
		args = append(args, *spec.Email)
	}
	if spec.Telegram != nil {
		parts = append(parts, "telegram = ?")
		args = append(args, *spec.Telegram)
	}
	return strings.Join(parts, " AND "), args
}

func fromSQLNull[T any](v sql.Null[T]) *T {
	if v.Valid {
		return &v.V
	}
	return nil
}

func toSQLNull[T any](v *T) sql.Null[T] {
	if v == nil {
		return sql.Null[T]{}
	}
	return sql.Null[T]{
		V:     *v,
		Valid: true,
	}
}
