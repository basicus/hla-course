package main

import (
	"context"
	"github.com/basicus/hla-course/log"
	"github.com/basicus/hla-course/service"
	eventconsumer "github.com/basicus/hla-course/service/event-consumer"
	eventproducer "github.com/basicus/hla-course/service/event-producer"
	"github.com/basicus/hla-course/service/monitoring"
	"github.com/basicus/hla-course/service/queue"
	"github.com/basicus/hla-course/service/rest"
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

	// REST Service
	restService, err := rest.New(cfg.Rest, logger, mon, &dbc, queueSrv)

	if err != nil {
		logger.WithError(err).Fatal("Cannot create rest service")
	}

	err = service.Setup(ctx, restService, "rest", g)
	if err != nil {
		logger.WithError(err).Fatal("Failed run rest service")
	}

	logger.Info("Running the service...")
	if err := g.Run(); err != nil {
		logger.WithError(err).Fatal("The service has been stopped with error")
	}
	logger.Info("The service is stopped")

}
