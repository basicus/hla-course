package log

// Config Конфигурация логгера
type Config struct {
	// Level уровень логирования
	Level string `env:"LOGGER_LEVEL,default=info"`
	// Timestamp отображать время вызова события
	DisableTimestamp bool `env:"LOGGER_DISABLE_TIMESTAMP"`
}
