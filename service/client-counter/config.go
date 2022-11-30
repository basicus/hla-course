package client_counter

type Config struct {
	ServiceChats string `env:"GRPC_CLIENT_COUNTER,default=localhost:9094"`
}
