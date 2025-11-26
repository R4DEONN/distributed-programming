package transport

import (
	"context"

	api "payment/api/server/paymentinternal"
)

func NewInternalAPI() api.PaymentInternalServiceServer {
	return &internalAPI{}
}

type internalAPI struct {
}

func (i *internalAPI) Ping(_ context.Context, _ *api.PingRequest) (*api.PingResponse, error) {
	return &api.PingResponse{
		Message: "pong",
	}, nil
}
