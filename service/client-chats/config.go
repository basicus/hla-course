package client_chats

type Config struct {
	ServiceChats string `env:"GRPC_CLIENT_CHATS,default=localhost:9092"`
}
