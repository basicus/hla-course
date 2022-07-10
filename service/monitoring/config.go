package monitoring

type Config struct {
	Listen string `env:"PROMETHEUS_LISTEN,default=localhost:8082"`
}
