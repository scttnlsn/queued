package queued

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestIterator(t *testing.T) {
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
