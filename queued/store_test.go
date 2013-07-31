package queued

import (
	"github.com/bmizerany/assert"
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	store := NewStore("./test1.db", true)
	defer store.Drop()

	item := NewItem("foo")
	store.Put(item)

	id := item.id

	result, ok := store.Get(id)

	assert.Equal(t, ok, true)
	assert.Equal(t, result, item)

	store.Remove(id)

	result, ok = store.Get(id)

	assert.Equal(t, ok, false)
}

func TestStoreConcurrency(t *testing.T) {
	store := NewStore("./test2.db", true)
	defer store.Drop()

	done := make(chan bool)

	put := func() {
		item := NewItem("foo")
		store.Put(item)
		done <- true
	}

	go put()
	go put()
	go put()

	<-done
	<-done
	<-done

	assert.Equal(t, store.LastId(), 3)
}

func TestStoreLoad(t *testing.T) {
	q := NewQueue("test_store_load")

	item := NewItem("foo")
	item.queue = q.Name

	temp := NewStore("./test3.db", true)
	temp.Put(item)
	temp.Close()

	store := NewStore("./test3.db", true)
	defer store.Drop()

	store.Load()

	_, ok := store.Get(item.id)
	assert.Equal(t, ok, true)

	ret, _ := q.Dequeue(time.Second, NilDuration)
	assert.Equal(t, ret.id, item.id)
}
