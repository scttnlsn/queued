package main

import (
	"./queued"
	"flag"
	"fmt"
	"runtime"
)

var config *queued.Config

func init() {
	config = queued.NewConfig()

	flag.UintVar(&config.Port, "port", 5353, "port on which to listen")
	flag.StringVar(&config.Auth, "auth", "", "HTTP basic auth password required for all requests")
	flag.StringVar(&config.DbPath, "db-path", "./queued.db", "the directory in which queue items will be persisted")
	flag.BoolVar(&config.Sync, "sync", true, "boolean indicating whether data should be synced to disk after every write")
}

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	s := queued.NewServer(config)
	s.Store.Load()

	err := s.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("main: %v", err))
	}

	fmt.Printf("Listening on http://localhost%s\n", s.Addr)

	shutdown := make(chan bool)
	<-shutdown
}
