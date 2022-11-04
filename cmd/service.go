package main

import (
	"context"
	"github.com/basicus/hla-course/log"
	"github.com/basicus/hla-course/service"
	client_auth "github.com/basicus/hla-course/service/client-auth"
	client_chats "github.com/basicus/hla-course/service/client-chats"
	eventconsumer "github.com/basicus/hla-course/service/event-consumer"
	eventproducer "github.com/basicus/hla-course/service/event-producer"
	grpc_auth "github.com/basicus/hla-course/service/grpc-auth"
	grpc_chats "github.com/basicus/hla-course/service/grpc-chats"
	"github.com/basicus/hla-course/service/monitoring"
	"github.com/basicus/hla-course/service/queue"
	"github.com/basicus/hla-course/service/rest"
	rest_chats "github.com/basicus/hla-course/service/rest-chats"
	wspusher "github.com/basicus/hla-course/service/wsclients"
	"github.com/basicus/hla-course/storage/mysql"
	"github.com/joeshaw/envdecode"
	"github.com/oklog/run"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

type config struct {
	Rest             rest.Config
	Mon              monitoring.Config
	Logger           log.Config
	Db               mysql.Config
	Queue            queue.Config
	Ws               wspusher.Config
	EvConsumerConfig eventconsumer.Config
	EvProducerConfig eventproducer.Config
	ClientAuth       client_auth.Config
	ClientChats      client_chats.Config
	GrpcAuth         grpc_auth.Config
	GrpcChats        grpc_chats.Config
	RestChats        rest_chats.Config
}

func main() {
	var cfg config
	if err := envdecode.StrictDecode(&cfg); err != nil {
		logrus.WithError(err).Fatal("Cannot decode config envs")
	}

	logger := log.New(cfg.Logger)

	ctx, cancel := context.WithCancel(log.WithContext(context.Background(), logrus.NewEntry(logger)))
	g := &run.Group{}
	{
		stop := make(chan os.Signal)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		g.Add(func() error {
			<-stop
			return nil
		}, func(error) {
			signal.Stop(stop)
			cancel()
			close(stop)
		})
	}

	// Database storage
	dbc, err := mysql.New(cfg.Db, logger)
	if err != nil {
		logger.WithError(err).Fatal("Cannot access to database")
	}

	// Database storage-chats
	dbcChats, err := mysql.NewChats(cfg.Db, logger)
	if err != nil {
		logger.WithError(err).Fatal("Cannot access to database")
	}

	// Monitoring
	mon, err := monitoring.New(cfg.Mon, logger)
	if err != nil {
		logger.WithError(err).Fatal("Cannot create monitoring service")
	}
	err = service.Setup(ctx, mon, "monitoring", g)
	if err != nil {
		logger.WithError(err).Fatal("Failed run monitoring service")
	}

	// EventProducer
	evProducer, err := eventproducer.New(cfg.EvProducerConfig, logger)
	if err != nil {
		logger.WithError(err).Fatal("Cannot create event producer service")
	}
	err = service.Setup(ctx, evProducer, "event_producer", g)
	if err != nil {
		logger.WithError(err).Fatal("Failed run event producer service")
	}

	// GRPC auth server
	grpcAuth, err := grpc_auth.New(cfg.GrpcAuth, &dbc, logger)
	err = service.Setup(ctx, grpcAuth, "grpc auth server", g)
	if err != nil {
		logger.WithError(err).Fatal("Failed run grpc auth server service")
	}

	// GRPC client for client auth
	clientAuth := client_auth.New(cfg.ClientAuth)
	clientAuth.Run(ctx)
	/*err = service.Setup(ctx, clientAuth, "grpc auth client", g)
	if err != nil {
		logger.WithError(err).Fatal("Failed run grpc auth client service")
	}*/

	// GRPC chats server
	grpcChats, err := grpc_chats.New(cfg.GrpcChats, &dbcChats, clientAuth, logger)
	err = service.Setup(ctx, grpcChats, "grpc chats server", g)
	if err != nil {
		logger.WithError(err).Fatal("Failed run grpc chats server service")
	}

	// GRPC client for chats
	clientChats := client_chats.New(cfg.ClientChats)
	clientChats.Run(ctx)
	/*	err = service.Setup(ctx, clientChats, "grpc chats client", g)
		if err != nil {
			logger.WithError(err).Fatal("Failed run grpc chats client service")
		}*/

	// Queue service
	queueSrv, err := queue.New(cfg.Queue, &dbc, logger, evProducer.PublishEvent)
	if err != nil {
		logger.WithError(err).Fatal("Cannot create queue service")
	}
	err = service.Setup(ctx, queueSrv, "queue", g)
	if err != nil {
		logger.WithError(err).Fatal("Failed run queue service")
	}

	// Websocket and message/post queue service
	wsSrv, err := wspusher.New(cfg.Ws, &dbc, logger)
	if err != nil {
		logger.WithError(err).Fatal("Cannot create websocket service")
	}
	err = service.Setup(ctx, wsSrv, "ws", g)
	if err != nil {
		logger.WithError(err).Fatal("Failed run websocket service")
	}

	// EventConsumer
	evConsumer, err := eventconsumer.New(cfg.EvConsumerConfig, wsSrv.SendEventToClient, logger)
	if err != nil {
		logger.WithError(err).Fatal("Cannot create event consumer service")
	}
	err = service.Setup(ctx, evConsumer, "event_consumer", g)
	if err != nil {
		logger.WithError(err).Fatal("Failed run event consumer service")
	}

	// REST Service main
	restService, err := rest.New(cfg.Rest, logger, mon, &dbc, queueSrv, nil, clientChats.Client)

	if err != nil {
		logger.WithError(err).Fatal("Cannot create rest service")
	}

	err = service.Setup(ctx, restService, "rest", g)
	if err != nil {
		logger.WithError(err).Fatal("Failed run rest service")
	}

	// REST Service chats
	restServiceChats, err := rest_chats.New(cfg.RestChats, logger, mon, &dbcChats, clientAuth.Client)

	if err != nil {
		logger.WithError(err).Fatal("Cannot create rest chat service")
	}

	err = service.Setup(ctx, restServiceChats, "rest-chats", g)
	if err != nil {
		logger.WithError(err).Fatal("Failed run rest chats service")
	}

	logger.Info("Running the service...")
	if err := g.Run(); err != nil {
		logger.WithError(err).Fatal("The service has been stopped with error")
	}
	logger.Info("The service is stopped")

}
