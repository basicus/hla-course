package queue

import "time"

type Config struct {
	Address                 string        `env:"REDIS_ADDRESS,default=localhost:6379"`
	UserName                string        `env:"REDIS_USERNAME"`
	Password                string        `env:"REDIS_PASSWORD,default=pass"`
	Database                int           `env:"REDIS_DATABASE,default=0"`
	PoolSize                int           `env:"REDIS_POOL_SIZE,default=5"`
	CleanPeriod             time.Duration `env:"QUEUE_CLEANUP_PERIOD,default=300s"`
	NumberConsumersForQueue int           `env:"CONSUMERS_PER_QUEUE,default=5"`
}
