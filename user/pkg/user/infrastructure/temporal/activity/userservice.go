package activity

import (
	"context"

	"github.com/google/uuid"

	appmodel "userservice/pkg/user/application/model"
	"userservice/pkg/user/application/service"
)

func NewUserServiceActivities(userService service.UserService) *UserServiceActivities {
	return &UserServiceActivities{userService: userService}
}

type UserServiceActivities struct {
	userService service.UserService
}

func (a *UserServiceActivities) FindUser(ctx context.Context, userID uuid.UUID) (appmodel.User, error) {
	return a.userService.FindUser(ctx, userID)
}

func (a *UserServiceActivities) SetUserStatus(ctx context.Context, userID uuid.UUID, status int) error {
	return a.userService.SetUserStatus(ctx, userID, status)
}
