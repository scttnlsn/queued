package queued

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/jmhodges/levigo"
	"strconv"
)

type Iterator struct {
	*levigo.Iterator
}

func NewIterator(db *levigo.DB) *Iterator {
	it := db.NewIterator(levigo.NewReadOptions())
	return &Iterator{it}
}

func (it *Iterator) Record() *Record {
	id, err := strconv.Atoi(string(it.Key()))
	if err != nil {
		panic(fmt.Sprintf("queued: Error loading db: %v", err))
	}

	value := it.Value()
	if value == nil {
		return nil
	}

	var record Record
	buf := bytes.NewBuffer(value)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&record)
	if err != nil {
		panic(fmt.Sprintf("queued.Store: Error decoding value: %v", err))
	}

	record.id = id
	return &record
}
