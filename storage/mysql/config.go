package mysql

type Config struct {
	DSN                  string `env:"DB_DSN,default=root:pass@tcp(localhost:3306)/project"`
	MaxOpenConnections   int    `env:"DB_MAX_OPEN_CONNECTIONS,default=5"`
	DSNro                string `env:"DB_DSN_RO"`
	MaxOpenConnectionsRo int    `env:"DB_RO_MAX_OPEN_CONNECTIONS,default=5"`
	RoDisable            bool   `env:"DB_RO_DISABLE,default=false"`
}
