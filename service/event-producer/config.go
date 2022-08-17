package eventproducer

type Config struct {
	QueueConnection    string `env:"RABBITMQ_CONNECTION,default=amqp://guest:guest@localhost:5672/"`
	QueueEventExchange string `env:"RABBITMQ_EVENT_EXCHANGE,default=Events"`
}
