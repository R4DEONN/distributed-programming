package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	applogging "gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/logging"
	"github.com/urfave/cli/v2"
)

const appID = "user"

func main() {
	logger := logging.NewJSONLogger(&logging.Config{AppName: appID})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = listenOSTermSignalsContext(ctx)

	app := cli.App{
		Name: appID,
		Commands: cli.Commands{
			migrate(logger),
			messageHandler(logger),
			workflowWorker(logger),
			service(logger),
		},
	}

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		logger.FatalError(err, "application stopped with error")
	}
}

func listenOSTermSignalsContext(ctx context.Context) context.Context {
	var cancelFunc context.CancelFunc
	ctx, cancelFunc = context.WithCancel(ctx)
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
		select {
		case <-ch:
			cancelFunc()
		case <-ctx.Done():
			return
		}
	}()
	return ctx
}

func graceCallback(ctx context.Context, logger applogging.Logger, gracePeriod time.Duration, callback func(ctx context.Context) error) {
	go func() {
		<-ctx.Done()
		graceCtx, cancel := context.WithTimeout(context.Background(), gracePeriod)
		defer cancel()

		err := callback(graceCtx)
		if err != nil {
			logger.Error(err, "graceful callback failed")
		}
	}()
}
