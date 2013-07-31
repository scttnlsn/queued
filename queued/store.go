package queued

import (
	"encoding/json"
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
	items map[int]*Item
	mutex sync.Mutex
	id    int
}

func NewStore(path string, sync bool) *Store {
	opts := levigo.NewOptions()
	opts.SetCreateIfMissing(true)

	db, err := levigo.Open(path, opts)
	if err != nil {
		panic(fmt.Sprintf("queued.Store: Unable to open db: %v", err))
	}

	store := &Store{
		path:  path,
		sync:  sync,
		db:    db,
		items: map[int]*Item{},
	}

	return store
}

func (s *Store) LastId() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	it := s.db.NewIterator(levigo.NewReadOptions())
	defer it.Close()

	it.SeekToLast()

	id := 0

	if it.Valid() {
		key, err := strconv.Atoi(string(it.Key()))
		if err != nil {
			panic(fmt.Sprintf("queued.Store: Error parsing last id from db: %v", err))
		}
		id = key
	}

	return id
}

func (s *Store) Get(id int) (*Item, bool) {
	item, ok := s.items[id]
	return item, ok
}

func (s *Store) Put(item *Item) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	item.id = s.id + 1

	key := []byte(fmt.Sprintf("%d", item.id))

	data := map[string]string{
		"value": item.value,
		"queue": item.queue,
	}

	bytes, err := json.Marshal(data)

	if err != nil {
		panic(fmt.Sprintf("queued.Store: Error marshalling item: %v", err))
	}

	wopts := levigo.NewWriteOptions()
	wopts.SetSync(s.sync)
	s.db.Put(wopts, key, bytes)

	s.items[item.id] = item
	s.id += 1
}

func (s *Store) Remove(id int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	key := []byte(fmt.Sprintf("%d", id))

	wopts := levigo.NewWriteOptions()
	wopts.SetSync(s.sync)
	s.db.Delete(wopts, key)

	delete(s.items, id)
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

func (s *Store) Load() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	it := s.db.NewIterator(levigo.NewReadOptions())
	defer it.Close()

	it.SeekToFirst()

	for it.Valid() {

		id, err := strconv.Atoi(string(it.Key()))
		if err != nil {
			panic(fmt.Sprintf("queued: Error loading db: %v", err))
		}

		data := make(map[string]string)

		err = json.Unmarshal(it.Value(), &data)
		if err != nil {
			panic(fmt.Sprintf("queued: Error loading db: %v", err))
		}

		item := NewItem(data["value"])
		item.id = id

		q := NewQueue(data["queue"])
		go q.Enqueue(item)

		s.items[id] = item
		s.id = id
		it.Next()
	}
}
