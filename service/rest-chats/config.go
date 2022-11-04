package rest_chats

type Config struct {
	Listen string `env:"CHATS_LISTEN_ADDRESS,default=localhost:8084"`
}
