package mysql

type Config struct {
	DSN                string `env:"DB_DSN,default=root:pass@tcp(localhost:3306)/project"`
	MaxOpenConnections int    `env:"DB_MAX_OPEN_CONNECTIONS,default=10"`
}
