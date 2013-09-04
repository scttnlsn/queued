package queued

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestStore(t *testing.T) {
	store := NewStore("./test1.db", true)
	defer store.Drop()

	assert.Equal(t, store.id, 0)

	record := NewRecord([]byte("foo"), "testqueue")

	err := store.Put(record)
	assert.Equal(t, err, nil)
	assert.Equal(t, record.id, 1)

	record, err = store.Get(1)
	assert.Equal(t, err, nil)
	assert.Equal(t, record.id, 1)
	assert.Equal(t, record.Value, []byte("foo"))
	assert.Equal(t, record.Queue, "testqueue")

	err = store.Remove(1)
	assert.Equal(t, err, nil)

	record, err = store.Get(1)
	assert.Equal(t, err, nil)
	assert.T(t, record == nil)
}

func TestStoreLoad(t *testing.T) {
	temp := NewStore("./test2.db", true)
	temp.Put(NewRecord([]byte("foo"), "testqueue"))
	temp.Put(NewRecord([]byte("bar"), "testqueue"))
	temp.Close()

	store := NewStore("./test2.db", true)
	defer store.Drop()

	assert.Equal(t, store.id, 2)
}

func TestStoreIterator(t *testing.T) {
	temp := NewStore("./test3.db", true)
	temp.Put(NewRecord([]byte("foo"), "testqueue"))
	temp.Put(NewRecord([]byte("bar"), "testqueue"))
	temp.Close()

	store := NewStore("./test3.db", true)
	defer store.Drop()

	it := store.Iterator()

	assert.T(t, it.Valid())

	one := it.Record()
	assert.Equal(t, one.id, 1)
	assert.Equal(t, one.Value, []byte("foo"))
	assert.Equal(t, one.Queue, "testqueue")

	it.Next()

	two := it.Record()
	assert.Equal(t, two.id, 2)
	assert.Equal(t, two.Value, []byte("bar"))
	assert.Equal(t, two.Queue, "testqueue")

	it.Next()

	assert.T(t, !it.Valid())
}
