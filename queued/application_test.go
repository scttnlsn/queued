package queued

import (
	"github.com/bmizerany/assert"
	"testing"
	"time"
)

func TestApplication(t *testing.T) {
	store := NewStore("./test1.db", true)
	defer store.Drop()

	app := NewApplication(store)

	assert.Equal(t, app.GetQueue("test"), app.GetQueue("test"))
	assert.NotEqual(t, app.GetQueue("test"), app.GetQueue("foobar"))

	record, err := app.Enqueue("test", []byte("foo"))

	assert.Equal(t, err, nil)
	assert.Equal(t, record.id, 1)
	assert.Equal(t, record.Value, []byte("foo"))
	assert.Equal(t, record.Queue, "test")

	info, err := app.Info("test", 1)

	assert.Equal(t, err, nil)
	assert.Equal(t, info["id"], 1)
	assert.Equal(t, info["queue"], "test")
	assert.Equal(t, info["dequeued"], false)

	record, err = app.Dequeue("test", NilDuration, NilDuration)

	assert.Equal(t, err, nil)
	assert.T(t, record != nil)
	assert.Equal(t, record.id, 1)
	assert.Equal(t, record.Value, []byte("foo"))
	assert.Equal(t, record.Queue, "test")

	ok, err := app.Complete("test", 1)
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, false)

	app.Enqueue("test", []byte("bar"))
	record, err = app.Dequeue("test", NilDuration, time.Millisecond)

	assert.Equal(t, err, nil)
	assert.T(t, record != nil)
	assert.Equal(t, record.id, 2)
	assert.Equal(t, record.Value, []byte("bar"))
	assert.Equal(t, record.Queue, "test")

	ok, err = app.Complete("test", 2)
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, true)

	ok, err = app.Complete("test", 2)
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, false)
}

func TestNewApplication(t *testing.T) {
	store := NewStore("./test2.db", true)
	defer store.Drop()

	store.Put(NewRecord([]byte("foo"), "test"))
	store.Put(NewRecord([]byte("bar"), "test"))
	store.Put(NewRecord([]byte("baz"), "another"))

	app := NewApplication(store)

	one, _ := app.Dequeue("test", NilDuration, NilDuration)
	assert.Equal(t, one.id, 1)
	assert.Equal(t, one.Value, []byte("foo"))

	two, _ := app.Dequeue("test", NilDuration, NilDuration)
	assert.Equal(t, two.id, 2)
	assert.Equal(t, two.Value, []byte("bar"))

	three, _ := app.Dequeue("another", NilDuration, NilDuration)
	assert.Equal(t, three.id, 3)
	assert.Equal(t, three.Value, []byte("baz"))
}
