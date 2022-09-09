package tarantool

type Config struct {
	DSN      string `env:"TARANTOOL_DSN,default=localhost:3301"`
	Password string `env:"TARANTOOL_PASSWORD"`
	UserName string `env:"TARANTOOL_USER"`
	Space    string `env:"TARANTOOL_SPACE,default=users"`
	Enable   bool   `env:"TARANTOOL_ENABLE,default=false"`
}
