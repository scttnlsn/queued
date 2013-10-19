package main

import (
	"flag"
	"fmt"
	"github.com/scttnlsn/queued/queued"
	"runtime"
)

var config *queued.Config

func init() {
	config = queued.NewConfig()

	flag.UintVar(&config.Port, "port", 5353, "port on which to listen")
	flag.StringVar(&config.Auth, "auth", "", "HTTP basic auth password required for all requests")
	flag.StringVar(&config.Store, "store", "leveldb", "the backend in which items will be stored (leveldb or memory)")
	flag.StringVar(&config.DbPath, "db-path", "./queued.db", "the directory in which queue items will be persisted (n/a for memory store)")
	flag.BoolVar(&config.Sync, "sync", true, "boolean indicating whether data should be synced to disk after every write (n/a for memory store)")
}

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	s := queued.NewServer(config)

	err := s.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("main: %v", err))
	}

	fmt.Printf("Listening on http://localhost%s\n", s.Addr)

	shutdown := make(chan bool)
	<-shutdown
}
