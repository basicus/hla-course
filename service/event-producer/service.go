package eventproducer

import (
	"context"
	"github.com/basicus/hla-course/log"
	"github.com/basicus/hla-course/model"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type Service struct {
	config Config
	log    *logrus.Logger
	conn   *amqp.Connection
	ch     *amqp.Channel
	close  chan struct{}
}

func New(config Config, log *logrus.Logger) (*Service, error) {
	return &Service{
		config: config,
		log:    log,
		conn:   nil,
		close:  make(chan struct{}),
	}, nil
}

func (s *Service) PublishEvent(ctx context.Context, userId int64, shardId string, event model.Event) error {
	s.log.Infof("request for publish event for user_id %d type %s body: %s", userId, event.GetType(), event.String())
	headers := make(map[string]interface{})
	const EventType = "type"
	const UserIdField = "user_id"
	if s.ch.IsClosed() {
		// Define channel
		ch, err := s.conn.Channel()
		s.log.Error("Channel closed opening it again.")
		if err != nil {
			s.log.WithError(err).Error("Failed to create channel")
			return err
		}
		s.ch = ch
	}
	headers[EventType] = event.GetType()
	headers[UserIdField] = userId
	_, err := s.ch.PublishWithDeferredConfirmWithContext(ctx,
		s.config.QueueEventExchange, // exchange
		shardId,                     // routing key by user shard
		false,                       // mandatory
		false,                       // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(event.String()),
			Headers:     headers,
		})
	return err
}

// Run Group task
func (s *Service) Run(ctx context.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("Start event producer")
	defer func() {
		logger.Info("Stop listening")
	}()

	conn, err := amqp.Dial(s.config.QueueConnection)

	if err != nil {
		logger.WithError(err).Error("Failed to connect server")
		return err
	}
	s.conn = conn
	logger.Info("Start connection established")

	// Define channel
	ch, err := s.conn.Channel()
	if err != nil {
		logger.WithError(err).Error("Failed to create channel")
		return err
	}
	s.ch = ch

	err = s.ch.ExchangeDeclare(
		s.config.QueueEventExchange, // Exchange
		"direct",                    // type
		true,                        // durable
		false,                       // auto-deleted
		false,                       // internal
		false,                       // no-wait
		nil,                         // arguments
	)

	if err != nil {
		logger.WithError(err).Error("Failed to create exchange")
		return err
	}

	// wait for closing
	<-s.close
	return nil

}

// Shutdown Group task gracefully shutdown
func (s *Service) Shutdown(ctx context.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("Closing connection and shutdown")
	_ = s.conn.Close()

	s.close <- struct{}{}
	return nil
}
