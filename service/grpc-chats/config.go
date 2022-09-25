package grpc_chats

type Config struct {
	Address string `env:"GRPC_CHATS_LISTEN,default=localhost:9092"`
}
