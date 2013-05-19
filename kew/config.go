package kew

type Config struct {
	Port   uint
	Auth   string
	DbPath string
}

func NewConfig() *Config {
	return &Config{5353, "", "./kew.db"}
}
