package grpc_auth

type Config struct {
	Address   string `env:"GRPC_AUTH_LISTEN,default=localhost:9093"`
	JwtSecret string `env:"JWT_SECRET,default=superpuper"` // TODO set require
}
