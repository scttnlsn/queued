package queued

import (
	"sync"
)

type Stats struct {
	Counters map[string]int
	mutex    sync.Mutex
}

func NewStats(counters map[string]int) *Stats {
	return &Stats{
		Counters: counters,
	}
}

func (s *Stats) Inc(field string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Counters[field] += 1
}

func (s *Stats) Dec(field string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Counters[field] -= 1
}

func (s *Stats) Get() map[string]int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.Counters
}
