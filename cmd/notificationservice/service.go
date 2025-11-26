package main

import (
	"context"
	"errors"
	"net"
	"net/http"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	libio "gitea.xscloud.ru/xscloud/golib/pkg/common/io"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"notificationservice/api/server/notificationinternal"
	appservice "notificationservice/pkg/notification/application/service"
	inframysql "notificationservice/pkg/notification/infrastructure/mysql"
	"notificationservice/pkg/notification/infrastructure/transport"
	"notificationservice/pkg/notification/infrastructure/transport/middlewares"
)

type serviceConfig struct {
	Service  Service  `envconfig:"service"`
	Database Database `envconfig:"database" required:"true"`
}

func service(logger logging.Logger) *cli.Command {
	return &cli.Command{
		Name:   "service",
		Before: migrateImpl(logger),
		Action: func(c *cli.Context) error {
			cnf, err := parseEnvs[serviceConfig]()
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

			libUoW := mysql.NewUnitOfWork(databaseConnectionPool, inframysql.NewRepositoryProvider)
			libLUow := mysql.NewLockableUnitOfWork(libUoW, mysql.NewLocker(databaseConnectionPool))

			notificationAPI := transport.NewNotificationInternalAPI(
				appservice.NewNotificationService(inframysql.NewLockableUnitOfWork(libLUow)),
			)

			errGroup := errgroup.Group{}
			errGroup.Go(func() error {
				listener, err := net.Listen("tcp", cnf.Service.GRPCAddress)
				if err != nil {
					return err
				}
				grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
					middlewares.NewGRPCLoggingMiddleware(logger),
				))
				notificationinternal.RegisterNotificationInternalServiceServer(grpcServer, notificationAPI)
				graceCallback(c.Context, logger, cnf.Service.GracePeriod, func(_ context.Context) error {
					grpcServer.GracefulStop()
					return nil
				})
				return grpcServer.Serve(listener)
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
