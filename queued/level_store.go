package queued

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"os"
	"strconv"
	"sync"
)

// Iterator

type LevelIterator struct {
	iterator.Iterator
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
	db    *leveldb.DB
	id    int
	mutex sync.Mutex
}

func NewLevelStore(path string, sync bool) *LevelStore {
	db, err := leveldb.OpenFile(path, &opt.Options{Flag: opt.OFCreateIfMissing})
	if err != nil {
		panic(fmt.Sprintf("queued.LevelStore: Unable to open db: %v", err))
	}

	id := 0

	it := db.NewIterator(&opt.ReadOptions{})
	defer it.Release()

	if it.Last() && it.Valid() {
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
	value, err := s.db.Get(key(id), &opt.ReadOptions{})
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		} else {
			return nil, err
		}
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

	wopts := &opt.WriteOptions{Flag: opt.WFSync}

	err = s.db.Put(key(id), buf.Bytes(), wopts)
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

	wopts := &opt.WriteOptions{Flag: opt.WFSync}

	return s.db.Delete(key(id), wopts)
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
	it := s.db.NewIterator(&opt.ReadOptions{})
	it.First()
	return &LevelIterator{it}
}

// Helpers

func key(id int) []byte {
	return []byte(fmt.Sprintf("%d", id))
}
