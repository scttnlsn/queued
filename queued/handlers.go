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
		send(w, http.StatusUnauthorized, Json{"error": "Unauthorized"})
		return false
	}

	return true
}

func (s *Server) EnqueueHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	value, err := ioutil.ReadAll(req.Body)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	record, err := s.App.Enqueue(params["queue"], value)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": err.Error()})
		return
	}

	w.Header().Set("Location", url(req, record))
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) DequeueHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	wait, err := Stod(req.URL.Query().Get("wait"), time.Second)
	if err != nil {
		send(w, http.StatusBadRequest, Json{"error": "Invalid wait parameter"})
		return
	}

	timeout, err := Stod(req.URL.Query().Get("timeout"), time.Second)
	if err != nil {
		send(w, http.StatusBadRequest, Json{"error": "Invalid timeout parameter"})
		return
	}

	record, err := s.App.Dequeue(params["queue"], wait, timeout)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": "Dequeue failed"})
		return
	}

	if record != nil {
		w.Header().Set("Location", url(req, record))
		w.Write(record.Value)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Server) InfoHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		send(w, http.StatusNotFound, Json{"error": "Item not found"})
		return
	}

	info, err := s.App.Info(params["queue"], id)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": "Failed to read item"})
		return
	}

	if info != nil {
		dequeued := "false"
		if info.dequeued {
			dequeued = "true"
		}

		w.Header().Set("X-Dequeued", dequeued)
		w.Write(info.value)
	} else {
		send(w, http.StatusNotFound, Json{"error": "Item not found"})
	}
}

func (s *Server) CompleteHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		send(w, http.StatusNotFound, Json{"error": "Item not found"})
		return
	}

	ok, err := s.App.Complete(params["queue"], id)
	if err != nil {
		send(w, http.StatusInternalServerError, Json{"error": "Complete failed"})
		return
	}

	if ok {
		w.WriteHeader(http.StatusNoContent)
	} else {
		send(w, http.StatusBadRequest, Json{"error": "Item not dequeued"})
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

func send(w http.ResponseWriter, code int, data Json) error {
	bytes, err := json.Marshal(data)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
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
