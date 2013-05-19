package kew

import (
	"fmt"
	"github.com/gorilla/mux"
	"net"
	"net/http"
)

type Server struct {
	Config *Config
	Router *mux.Router
	Store  *Store
	Addr   string
}

func NewServer(config *Config) *Server {
	router := mux.NewRouter()
	store := NewStore(config.DbPath)
	addr := fmt.Sprintf(":%d", config.Port)

	s := &Server{config, router, store, addr}

	s.HandleFunc("/{queue}", s.EnqueueHandler).Methods("POST")
	s.HandleFunc("/{queue}/head", s.DequeueHandler).Methods("GET")
	s.HandleFunc("/{queue}/{id}", s.InfoHandler).Methods("GET")
	s.HandleFunc("/{queue}/{id}", s.CompleteHandler).Methods("DELETE")

	return s
}

func (s *Server) HandleFunc(route string, fn func(http.ResponseWriter, *http.Request)) *mux.Route {
	return s.Router.HandleFunc(route, func(w http.ResponseWriter, req *http.Request) {
		if ok := s.BeforeHandler(w, req); ok {
			fn(w, req)
		}
	})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.Router.ServeHTTP(w, req)
}

func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.Addr)

	if err != nil {
		return err
	}

	srv := http.Server{Handler: s}
	go srv.Serve(listener)

	return nil
}
