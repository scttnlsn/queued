package queued

import (
	"strconv"
	"testing"
	"time"

	"github.com/bmizerany/assert"
)

func TestApplication(t *testing.T) {
	store := NewLevelStore("./test1.db", true)
	defer store.Drop()

	app := NewApplication(store)

	assert.Equal(t, app.GetQueue("test"), app.GetQueue("test"))
	assert.NotEqual(t, app.GetQueue("test"), app.GetQueue("foobar"))

	record, err := app.Enqueue("test", []byte("foo"), "")

	assert.Equal(t, err, nil)
	assert.Equal(t, record.Id, 1)
	assert.Equal(t, record.Value, []byte("foo"))
	assert.Equal(t, record.Queue, "test")

	stats := app.Stats("test")

	assert.Equal(t, stats["enqueued"], 1)
	assert.Equal(t, stats["dequeued"], 0)
	assert.Equal(t, stats["depth"], 1)
	assert.Equal(t, stats["timeouts"], 0)

	info, err := app.Info("test", 1)

	assert.Equal(t, err, nil)
	assert.Equal(t, info.record.Value, []byte("foo"))
	assert.Equal(t, info.dequeued, false)

	record, err = app.Dequeue("test", NilDuration, NilDuration)

	assert.Equal(t, err, nil)
	assert.T(t, record != nil)
	assert.Equal(t, record.Id, 1)
	assert.Equal(t, record.Value, []byte("foo"))
	assert.Equal(t, record.Queue, "test")

	ok, err := app.Complete("test", 1)
	assert.Equal(t, err, nil)
	assert.Equal(t, ok, false)

	app.Enqueue("test", []byte("bar"), "")
	record, err = app.Dequeue("test", NilDuration, time.Millisecond)

	assert.Equal(t, err, nil)
	assert.T(t, record != nil)
	assert.Equal(t, record.Id, 2)
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
	store := NewLevelStore("./test2.db", true)
	defer store.Drop()

	store.Put(NewRecord([]byte("foo"), "test"))
	store.Put(NewRecord([]byte("bar"), "test"))
	store.Put(NewRecord([]byte("baz"), "another"))

	app := NewApplication(store)

	one, _ := app.Dequeue("test", NilDuration, NilDuration)
	assert.Equal(t, one.Id, 1)
	assert.Equal(t, one.Value, []byte("foo"))

	two, _ := app.Dequeue("test", NilDuration, NilDuration)
	assert.Equal(t, two.Id, 2)
	assert.Equal(t, two.Value, []byte("bar"))

	three, _ := app.Dequeue("another", NilDuration, NilDuration)
	assert.Equal(t, three.Id, 3)
	assert.Equal(t, three.Value, []byte("baz"))
}

func BenchmarkSmallQueue(b *testing.B) {
	store := NewLevelStore("./bench1.db", true)
	defer store.Drop()
	app := NewApplication(store)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Enqueue("test", []byte(strconv.Itoa(i)), "")
	}
}

func BenchmarkSmallDequeue(b *testing.B) {
	store := NewLevelStore("./bench2.db", true)
	defer store.Drop()
	app := NewApplication(store)
	for i := 0; i < b.N; i++ {
		app.Enqueue("test", []byte(strconv.Itoa(i)), "")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Dequeue("test", NilDuration, NilDuration)
	}
}

var testValue = []byte(`{ "glossary": { "title": "example glossary", "GlossDiv": { "title": "S", "GlossList": { "GlossEntry": { "ID": "SGML", "SortAs": "SGML", "GlossTerm": "Standard Generalized Markup Language", "Acronym": "SGML", "Abbrev": "ISO 8879:1986", "GlossDef": { "para": "A meta-markup language, used to create markup languages such as DocBook.", "GlossSeeAlso": ["GML", "XML"] }, "GlossSee": "markup" } } } } }`)

func BenchmarkQueue(b *testing.B) {
	store := NewLevelStore("./bench1.db", true)
	defer store.Drop()
	app := NewApplication(store)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Enqueue("test", testValue, "")
	}

}

func BenchmarkDequeue(b *testing.B) {
	store := NewLevelStore("./bench2.db", true)
	defer store.Drop()
	app := NewApplication(store)
	for i := 0; i < b.N; i++ {
		app.Enqueue("test", testValue, "")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Dequeue("test", NilDuration, NilDuration)
	}
}
