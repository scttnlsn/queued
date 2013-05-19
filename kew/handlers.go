package kew

import (
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (s *Server) BeforeHandler(w http.ResponseWriter, req *http.Request) bool {
	if ok := auth(req, s.Config); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		send(w, Json{"error": "Unauthorized"})
		return false
	}

	return true
}

func (s *Server) EnqueueHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	q := NewQueue(params["queue"])

	value, err := ioutil.ReadAll(req.Body)
	if err != nil {
		send(w, Json{"error": err.Error()})
		return
	}

	item := NewItem(string(value))
	item.queue = q.Name

	s.Store.Put(item)

	go q.Enqueue(item)

	w.WriteHeader(http.StatusCreated)
	send(w, Json{"id": item.id, "value": item.value})
}

func (s *Server) DequeueHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	q := NewQueue(params["queue"])

	wait, err := Stod(req.URL.Query().Get("wait"), time.Second)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		send(w, Json{"error": "Invalid wait parameter"})
		return
	}

	timeout, err := Stod(req.URL.Query().Get("timeout"), time.Second)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		send(w, Json{"error": "Invalid timeout parameter"})
		return
	}

	if item, ok := q.Dequeue(wait, timeout); ok {
		if timeout == NilDuration {
			s.Store.Remove(item.id)
		}

		send(w, Json{"id": item.id, "value": item.value})
	} else {
		w.WriteHeader(http.StatusNotFound)
		send(w, Json{"id": nil})
	}
}

func (s *Server) InfoHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		send(w, Json{"error": "Item not found"})
		return
	}

	item, ok := s.Store.Get(id)

	if !ok || item.queue != params["queue"] {
		w.WriteHeader(http.StatusNotFound)
		send(w, Json{"error": "Item not found"})
		return
	}

	send(w, Json{"id": item.id, "value": item.value, "dequeued": item.dequeued})
}

func (s *Server) CompleteHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		send(w, Json{"error": "Item not found"})
		return
	}

	item, ok := s.Store.Get(id)

	if !ok || item.queue != params["queue"] {
		w.WriteHeader(http.StatusNotFound)
		send(w, Json{"error": "Item not found"})
		return
	}

	ok = item.Complete()

	if ok {
		s.Store.Remove(item.id)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	send(w, Json{"ok": ok})
}

// Helpers

type Json map[string]interface{}

func Stod(val string, scale ...time.Duration) (time.Duration, error) {
	duration := NilDuration

	if val != "" {
		n, err := strconv.Atoi(val)

		if err != nil {
			return duration, err
		} else {
			duration = time.Duration(n)

			if len(scale) == 1 {
				duration *= scale[0]
			}
		}
	}

	return duration, nil
}

func send(w http.ResponseWriter, data Json) error {
	bytes, err := json.Marshal(data)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)

	return nil
}

func auth(req *http.Request, config *Config) bool {
	if config.Auth == "" {
		return true
	}

	s := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 || s[0] != "Basic" {
		return false
	}

	base, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(base), ":", 2)
	if len(pair) != 2 {
		return false
	}

	password := pair[1]
	if config.Auth != password {
		return false
	}

	return true
}
