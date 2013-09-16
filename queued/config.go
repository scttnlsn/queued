package queued

import (
	"fmt"
)

type Config struct {
	Port   uint
	Auth   string
	Store  string
	DbPath string
	Sync   bool
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) CreateStore() Store {
	if c.Store == "leveldb" {
		return NewLevelStore(c.DbPath, c.Sync)
	} else if c.Store == "memory" {
		return NewMemoryStore()
	} else {
		panic(fmt.Sprintf("queued.Config: Invalid store: %s", c.Store))
	}
}
