package main

import (
	"./kew"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
)

func main() {
	flag.Parse()
	path := flag.Arg(0)

	runtime.GOMAXPROCS(runtime.NumCPU())

	config := kew.NewConfig()

	// Load config
	if path != "" {
		file, err := ioutil.ReadFile(path)
		if err != nil {
			panic(fmt.Sprintf("kew: Error reading config file: %v", err))
			os.Exit(1)
		}

		err = json.Unmarshal(file, &config)
		if err != nil {
			panic(fmt.Sprintf("kew: Error parsing config file: %v", err))
		}
	}

	s := kew.NewServer(config)
	s.Store.Load()

	err := s.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("main: %v", err))
	}

	fmt.Printf("Listening on http://localhost%s\n", s.Addr)

	shutdown := make(chan bool)
	<-shutdown
}
