package wspusher

type Config struct {
	Listen string `env:"WS_LISTEN_ADDRESS,default=localhost:8081"`
}
