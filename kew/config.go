package kew

type Config struct {
	Port   uint
	Auth   string
	DbPath string
	Sync   bool
}

func NewConfig() *Config {
	return &Config{}
}
