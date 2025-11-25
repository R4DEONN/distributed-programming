package transport

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"userservice/api/server/userpublicapi"
	appmodel "userservice/pkg/user/application/model"
	"userservice/pkg/user/application/query"
	"userservice/pkg/user/application/service"
)

func NewUserInternalAPI(
	userQueryService query.UserQueryService,
	userService service.UserService,
) userpublicapi.UserPublicAPIServer {
	return &userInternalAPI{
		userQueryService: userQueryService,
		userService:      userService,
	}
}

type userInternalAPI struct {
	userQueryService query.UserQueryService
	userService      service.UserService

	userpublicapi.UnimplementedUserPublicAPIServer
}

func (u userInternalAPI) StoreUser(ctx context.Context, request *userpublicapi.StoreUserRequest) (*userpublicapi.StoreUserResponse, error) {
	var (
		userID uuid.UUID
		err    error
	)
	if request.UserID != "" {
		userID, err = uuid.Parse(request.UserID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", request.UserID)
		}
	}

	userID, err = u.userService.StoreUser(ctx, appmodel.User{
		UserID:   userID,
		Login:    request.Login,
		Email:    request.Email,
		Telegram: request.Telegram,
	})
	if err != nil {
		return nil, err
	}

	return &userpublicapi.StoreUserResponse{
		UserID: userID.String(),
	}, nil
}

func (u userInternalAPI) FindUser(ctx context.Context, request *userpublicapi.FindUserRequest) (*userpublicapi.FindUserResponse, error) {
	userID, err := uuid.Parse(request.UserID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", request.UserID)
	}
	user, err := u.userQueryService.FindUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user %q not found", request.UserID)
	}
	return &userpublicapi.FindUserResponse{
		UserID:   userID.String(),
		Status:   userpublicapi.UserStatus(user.Status), // nolint:gosec
		Login:    user.Login,
		Email:    user.Email,
		Telegram: user.Telegram,
	}, nil
}
