package config

type Config struct {
	Port        string `env:"PORT"          envDefault:"8080"`
	DatabaseURL string `env:"DATABASE_URL,required"`
	LogLevel    string `env:"LOG_LEVEL"     envDefault:"info"`
}
