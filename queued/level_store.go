package queued

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/jmhodges/levigo"
	"os"
	"strconv"
	"sync"
)

// Iterator

type LevelIterator struct {
	*levigo.Iterator
}

func (it *LevelIterator) NextRecord() (*Record, bool) {
	if !it.Valid() {
		return nil, false
	}

	id, err := strconv.Atoi(string(it.Key()))
	if err != nil {
		panic(fmt.Sprintf("queued.LevelIterator: Error loading db: %v", err))
	}

	value := it.Value()
	if value == nil {
		return nil, false
	}

	var record Record
	buf := bytes.NewBuffer(value)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&record)
	if err != nil {
		panic(fmt.Sprintf("queued.LevelIterator: Error decoding value: %v", err))
	}

	record.id = id

	it.Next()
	return &record, true
}

// Store

type LevelStore struct {
	path  string
	sync  bool
	db    *levigo.DB
	id    int
	mutex sync.Mutex
}

func NewLevelStore(path string, sync bool) *LevelStore {
	opts := levigo.NewOptions()
	opts.SetCreateIfMissing(true)

	db, err := levigo.Open(path, opts)
	if err != nil {
		panic(fmt.Sprintf("queued.LevelStore: Unable to open db: %v", err))
	}

	id := 0

	it := db.NewIterator(levigo.NewReadOptions())
	defer it.Close()

	it.SeekToLast()
	if it.Valid() {
		id, err = strconv.Atoi(string(it.Key()))
		if err != nil {
			panic(fmt.Sprintf("queued.LevelStore: Error loading db: %v", err))
		}
	}

	return &LevelStore{
		id:   id,
		path: path,
		sync: sync,
		db:   db,
	}
}

func (s *LevelStore) Get(id int) (*Record, error) {
	ropts := levigo.NewReadOptions()

	value, err := s.db.Get(ropts, key(id))
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}

	var record Record
	buf := bytes.NewBuffer(value)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&record)
	if err != nil {
		panic(fmt.Sprintf("queued.LevelStore: Error decoding value: %v", err))
	}

	record.id = id

	return &record, nil
}

func (s *LevelStore) Put(record *Record) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := s.id + 1

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(record)
	if err != nil {
		panic(fmt.Sprintf("queued.LevelStore: Error encoding record: %v", err))
	}

	wopts := levigo.NewWriteOptions()
	wopts.SetSync(s.sync)

	err = s.db.Put(wopts, key(id), buf.Bytes())
	if err != nil {
		return err
	}

	record.id = id
	s.id = id

	return nil
}

func (s *LevelStore) Remove(id int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	wopts := levigo.NewWriteOptions()
	wopts.SetSync(s.sync)
	return s.db.Delete(wopts, key(id))
}

func (s *LevelStore) Close() {
	s.db.Close()
}

func (s *LevelStore) Drop() {
	s.Close()

	err := os.RemoveAll(s.path)
	if err != nil {
		panic(fmt.Sprintf("queued.LevelStore: Error dropping db: %v", err))
	}
}

func (s *LevelStore) Iterator() Iterator {
	it := s.db.NewIterator(levigo.NewReadOptions())
	it.SeekToFirst()
	return &LevelIterator{it}
}

// Helpers

func key(id int) []byte {
	return []byte(fmt.Sprintf("%d", id))
}
