package grpc_counter

type Config struct {
	Address string `env:"GRPC_COUNTER_LISTEN,default=localhost:9094"`
}
