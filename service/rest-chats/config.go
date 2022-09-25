package rest_chats

type Config struct {
	Listen string `env:"LISTEN_ADDRESS,default=localhost:8084"`
}
