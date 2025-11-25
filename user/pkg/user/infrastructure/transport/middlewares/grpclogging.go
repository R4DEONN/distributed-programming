package middlewares

import (
	"context"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"google.golang.org/grpc"
)

func NewGRPCLoggingMiddleware(logger logging.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		fields := logging.Fields{
			"args":     req,
			"duration": time.Since(start).String(),
			"method":   info.FullMethod,
		}

		l := logger.WithFields(fields)
		if err != nil {
			l.Error(err, "call failed")
		} else {
			l.Info("call finished")
		}
		return resp, err
	}
}
