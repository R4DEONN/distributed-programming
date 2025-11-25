package query

import (
	"context"
	"database/sql"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	appmodel "userservice/pkg/user/application/model"
	"userservice/pkg/user/application/query"
	"userservice/pkg/user/domain/model"
)

func NewUserQueryService(client mysql.ClientContext) query.UserQueryService {
	return &userQueryService{
		client: client,
	}
}

type userQueryService struct {
	client mysql.ClientContext
}

func (u *userQueryService) FindUser(ctx context.Context, userID uuid.UUID) (*appmodel.User, error) {
	user := struct {
		UserID   uuid.UUID        `db:"user_id"`
		Status   int              `db:"status"`
		Login    string           `db:"login"`
		Email    sql.Null[string] `db:"email"`
		Telegram sql.Null[string] `db:"telegram"`
	}{}

	err := u.client.GetContext(
		ctx,
		&user,
		`SELECT user_id, status, login, email, telegram FROM user WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrUserNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &appmodel.User{
		UserID:   user.UserID,
		Status:   user.Status,
		Login:    user.Login,
		Email:    fromSQLNull(user.Email),
		Telegram: fromSQLNull(user.Telegram),
	}, nil
}

func fromSQLNull[T any](v sql.Null[T]) *T {
	if v.Valid {
		return &v.V
	}
	return nil
}
