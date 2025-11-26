package transport

import (
	"context"

	api "product/api/server/productinternal"
)

func NewInternalAPI() api.ProductInternalServiceServer {
	return &internalAPI{}
}

type internalAPI struct {
}

func (i *internalAPI) Ping(_ context.Context, _ *api.PingRequest) (*api.PingResponse, error) {
	return &api.PingResponse{
		Message: "pong",
	}, nil
}
