package transport

import (
	"context"

	api "notification/api/server/notificationinternal"
)

func NewInternalAPI() api.NotificationInternalServiceServer {
	return &internalAPI{}
}

type internalAPI struct {
}

func (i *internalAPI) Ping(_ context.Context, _ *api.PingRequest) (*api.PingResponse, error) {
	return &api.PingResponse{
		Message: "pong",
	}, nil
}
