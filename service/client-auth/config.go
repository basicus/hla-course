package client_auth

type Config struct {
	ServiceAuth string `env:"GRPC_CLIENT_AUTH,default=localhost:9093"`
}
