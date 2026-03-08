package config

type Config struct {
	DBHost string
	DBPort int
}

func Load() *Config {
	return &Config{
		DBHost: "localhost",
		DBPort: 3306,
	}
}
