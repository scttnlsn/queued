package queued

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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

	value, err := ioutil.ReadAll(req.Body)
	if err != nil {
		send(w, Json{"error": err.Error()})
		return
	}

	record, err := s.App.Enqueue(params["queue"], value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		send(w, Json{"error": "Enqueue failed"})
		return
	}

	w.Header().Set("Location", url(req, record))
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) DequeueHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

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

	record, err := s.App.Dequeue(params["queue"], wait, timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		send(w, Json{"error": "Dequeue failed"})
		return
	}

	if record != nil {
		w.Header().Set("Location", url(req, record))
		fmt.Fprintf(w, "%s", record.Value)
	} else {
		w.WriteHeader(http.StatusNotFound)
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

	info, err := s.App.Info(params["queue"], id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		send(w, Json{"error": "Failed to read item"})
		return
	}

	if info != nil {
		dequeued := "false"
		if info.dequeued {
			dequeued = "true"
		}

		w.Header().Set("X-Dequeued", dequeued)
		fmt.Fprintf(w, "%s", info.value)
	} else {
		w.WriteHeader(http.StatusNotFound)
		send(w, Json{"error": "Item not found"})
		return
	}
}

func (s *Server) CompleteHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		send(w, Json{"error": "Item not found"})
		return
	}

	ok, err := s.App.Complete(params["queue"], id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		send(w, Json{"error": "Complete failed"})
		return
	}

	if ok {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		send(w, Json{"error": "Item not dequeued"})
	}
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

func url(req *http.Request, record *Record) string {
	return fmt.Sprintf("http://%s/%s/%d", req.Host, record.Queue, record.id)
}
