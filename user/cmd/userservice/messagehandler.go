package main

import (
	"errors"
	"net/http"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	libio "gitea.xscloud.ru/xscloud/golib/pkg/common/io"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/outbox"
	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"userservice/pkg/user/infrastructure/integrationevent"
	"userservice/pkg/user/infrastructure/temporal"
)

type messageHandlerConfig struct {
	Service  Service  `envconfig:"service"`
	Database Database `envconfig:"database" required:"true"`
	AMQP     AMQP     `envconfig:"amqp" required:"true"`
	Temporal Temporal `envconfig:"temporal" required:"true"`
}

func messageHandler(logger logging.Logger) *cli.Command {
	return &cli.Command{
		Name:   "message-handler",
		Before: migrateImpl(logger),
		Action: func(c *cli.Context) error {
			cnf, err := parseEnvs[messageHandlerConfig]()
			if err != nil {
				return err
			}

			closer := libio.NewMultiCloser()
			defer func() {
				err = errors.Join(err, closer.Close())
			}()

			databaseConnector, err := newDatabaseConnector(cnf.Database)
			if err != nil {
				return err
			}
			closer.AddCloser(databaseConnector)
			databaseConnectionPool := mysql.NewConnectionPool(databaseConnector.TransactionalClient())

			temporalClient, err := temporal.NewClient(logger, cnf.Temporal.Host)
			if err != nil {
				return err
			}
			closer.AddCloser(libio.CloserFunc(func() error {
				temporalClient.Close()
				return nil
			}))
			workflowService := temporal.NewWorkflowService(temporalClient)

			amqpConnection := newAMQPConnection(cnf.AMQP, logger)
			queueConfig := &amqp.QueueConfig{
				Name:    integrationevent.QueueName,
				Durable: true,
			}
			bindConfig := &amqp.BindConfig{
				QueueName:    integrationevent.QueueName,
				ExchangeName: integrationevent.ExchangeName,
				RoutingKeys:  []string{integrationevent.RoutingKeyPrefix + "#"},
			}
			amqpEventProducer := amqpConnection.Producer(
				&amqp.ExchangeConfig{
					Name:    integrationevent.ExchangeName,
					Kind:    integrationevent.ExchangeKind,
					Durable: true,
				},
				queueConfig,
				bindConfig,
			)
			amqpTransport := integrationevent.NewAMQPTransport(logger, workflowService)
			amqpConnection.Consumer(
				c.Context,
				amqpTransport.Handler(),
				queueConfig,
				bindConfig,
				&amqp.QoSConfig{
					PrefetchCount: 100,
				},
			)
			err = amqpConnection.Start()
			if err != nil {
				return err
			}
			closer.AddCloser(libio.CloserFunc(func() error {
				return amqpConnection.Stop()
			}))

			outboxEventHandler := outbox.NewEventHandler(outbox.EventHandlerConfig{
				TransportName:  integrationevent.TransportName,
				Transport:      integrationevent.NewOutboxTransport(logger, amqpEventProducer),
				ConnectionPool: databaseConnectionPool,
				Logger:         logger,
			})

			errGroup := errgroup.Group{}
			errGroup.Go(func() error {
				return outboxEventHandler.Start(c.Context)
			})

			errGroup.Go(func() error {
				router := mux.NewRouter()
				registerHealthcheck(router)
				// nolint:gosec
				server := http.Server{
					Addr:    cnf.Service.HTTPAddress,
					Handler: router,
				}
				graceCallback(c.Context, logger, cnf.Service.GracePeriod, server.Shutdown)
				return server.ListenAndServe()
			})

			return errGroup.Wait()
		},
	}
}
