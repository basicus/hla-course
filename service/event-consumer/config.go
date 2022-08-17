package eventconsumer

type Config struct {
	QueueConnection    string `env:"RABBITMQ_CONNECTION,default=amqp://guest:guest@localhost:5672/"`
	QueueRoutingKey    string `env:"QUEUE_RKEY,default=00000"`
	QueueEventExchange string `env:"RABBITMQ_EVENT_EXCHANGE,default=Events"`
	QueueEventPrefix   string `env:"RABBITMQ_EVENT_QUEUE,default=Events"`
}
