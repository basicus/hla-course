package eventconsumer

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
	queue  amqp.Queue
	close  chan struct{}
	msgs   <-chan amqp.Delivery
	send   func(event model.WsEvent) error
}

func New(config Config, send func(event model.WsEvent) error, log *logrus.Logger) (*Service, error) {
	return &Service{
		config: config,
		log:    log,
		conn:   nil,
		close:  make(chan struct{}),
		send:   send,
	}, nil
}

func (s *Service) ConsumeEvents() {
	s.log.Infof("Start consuming loop")
	const EventType = "type"
	const UserIdField = "user_id"
	for d := range s.msgs {
		evType := ""
		variable, ok := d.Headers[EventType].(string)
		if ok {
			evType = variable
		}
		s.log.Infof("Received event for user_id %s type %s : %s", d.UserId, evType, d.Body)
		var userId int64
		vUserId, ok := d.Headers[UserIdField].(int64)
		if ok {
			userId = vUserId
		}
		err := s.send(model.WsEvent{
			UserId:  userId,
			Message: string(d.Body),
		})
		if err != nil {
			s.log.Errorf("Send event to websocket done: %s", err)
		} else {
			s.log.Infof("Send event to websocket done")
		}

		_ = d.Ack(false)
	}
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

	// Declare queue
	queue, err := ch.QueueDeclare(
		s.config.QueueEventPrefix+"_"+s.config.QueueRoutingKey, // name
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.WithError(err).Error("Failed to define queue")
		return err
	}
	s.queue = queue

	// Declare bindings by routekey
	err = ch.QueueBind(
		s.queue.Name,                // queue name
		s.config.QueueRoutingKey,    // routing key
		s.config.QueueEventExchange, // exchange
		false,
		nil)
	if err != nil {
		logger.WithError(err).Error("Failed to set binding")
		return err
	}

	// Consume
	msgs, err := s.ch.Consume(
		s.config.QueueEventPrefix+"_"+s.config.QueueRoutingKey, // name
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	s.msgs = msgs

	// Run consumer
	go s.ConsumeEvents()
	// wait for closing
	<-s.close
	return nil

}

// Shutdown Group task gracefull shutdown
func (s *Service) Shutdown(ctx context.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("Closing connection and shutdown")
	s.conn.Close()
	s.close <- struct{}{}
	return nil
}
