// +build use_goleveldb

package queued

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	leveldb_iterator "github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Iterator

type LevelIterator struct {
	leveldb_iterator.Iterator
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

	record.Id = id

	it.Next()
	return &record, true
}

// Store

type LevelStore struct {
	path  string
	sync  bool
	db    *leveldb.DB
	id    int
	mutex sync.Mutex
}

func NewLevelStore(path string, sync bool) *LevelStore {
	opts := &opt.Options{
		Filter:         filter.NewBloomFilter(10),
		ErrorIfMissing: false,
	}
	db, err := leveldb.OpenFile(path, opts)
	if err != nil {
		panic(fmt.Sprintf("queued.LevelStore: Unable to open db: %v", err))
	}

	id := 0

	iter := db.NewIterator(nil, nil)
	iter.Last()
	if iter.Valid() {
		id, err = strconv.Atoi(string(iter.Key()))
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
	value, err := s.db.Get(key(id), nil)
	if err == util.ErrNotFound {
		return nil, nil
	}
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

	record.Id = id

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

	err = s.db.Put(key(id), buf.Bytes(), &opt.WriteOptions{Sync: s.sync})
	if err != nil {
		return err
	}

	record.Id = id
	s.id = id

	return nil
}

func (s *LevelStore) Remove(id int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.db.Delete(key(id), &opt.WriteOptions{Sync: s.sync})
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
	it := s.db.NewIterator(nil, nil)
	it.First()
	return &LevelIterator{it}
}

// Helpers

func key(id int) []byte {
	return []byte(fmt.Sprintf("%d", id))
}
