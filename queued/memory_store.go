package queued

import (
	"sync"
)

// Iterator

type MemoryIterator struct {
}

func (it *MemoryIterator) NextRecord() (*Record, bool) {
	return nil, false
}

// Store

type MemoryStore struct {
	id      int
	records map[int]*Record
	mutex   sync.Mutex
}

func NewMemoryStore() *MemoryStore {
	records := make(map[int]*Record)

	return &MemoryStore{
		id:      0,
		records: records,
	}
}

func (s *MemoryStore) Get(id int) (*Record, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if record, ok := s.records[id]; ok {
		return record, nil
	} else {
		return nil, nil
	}
}

func (s *MemoryStore) Put(record *Record) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	record.Id = s.id + 1
	s.records[record.Id] = record
	s.id = record.Id
	return nil
}

func (s *MemoryStore) Remove(id int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.records, id)
	return nil
}

func (s *MemoryStore) Iterator() Iterator {
	return &MemoryIterator{}
}

func (s *MemoryStore) Drop() {
	s.records = make(map[int]*Record)
}
