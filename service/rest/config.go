package rest

type Config struct {
	Listen     string `env:"LISTEN_ADDRESS,default=localhost:8080"`
	JwtSecret  string `env:"JWT_SECRET,default=superpuper"`
	PostsLimit int64  `env:"FRIENDS_POSTS_LIMIT,default=1000"`
}
