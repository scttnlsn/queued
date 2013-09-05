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

type Store struct {
	path  string
	sync  bool
	db    *levigo.DB
	id    int
	mutex sync.Mutex
}

func NewStore(path string, sync bool) *Store {
	opts := levigo.NewOptions()
	opts.SetCreateIfMissing(true)

	db, err := levigo.Open(path, opts)
	if err != nil {
		panic(fmt.Sprintf("queued.Store: Unable to open db: %v", err))
	}

	id := 0

	it := db.NewIterator(levigo.NewReadOptions())
	defer it.Close()

	it.SeekToLast()
	if it.Valid() {
		id, err = strconv.Atoi(string(it.Key()))
		if err != nil {
			panic(fmt.Sprintf("queued: Error loading db: %v", err))
		}
	}

	store := &Store{
		id:   id,
		path: path,
		sync: sync,
		db:   db,
	}

	return store
}

func (s *Store) Get(id int) (*Record, error) {
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
		panic(fmt.Sprintf("queued.Store: Error decoding value: %v", err))
	}

	record.id = id

	return &record, nil
}

func (s *Store) Put(record *Record) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := s.id + 1

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(record)
	if err != nil {
		panic(fmt.Sprintf("queued.Store: Error encoding record: %v", err))
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

func (s *Store) Remove(id int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	wopts := levigo.NewWriteOptions()
	wopts.SetSync(s.sync)
	return s.db.Delete(wopts, key(id))
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) Drop() {
	s.Close()

	err := os.RemoveAll(s.path)
	if err != nil {
		panic(fmt.Sprintf("queued.Store: Error dropping db: %v", err))
	}
}

func (s *Store) Iterator() *Iterator {
	it := NewIterator(s.db)
	it.SeekToFirst()
	return it
}

// Helpers

func key(id int) []byte {
	return []byte(fmt.Sprintf("%d", id))
}
