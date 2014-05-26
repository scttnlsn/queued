package queued

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()
	defer store.Drop()

	assert.Equal(t, store.id, 0)

	record := NewRecord([]byte("foo"), "testqueue")

	err := store.Put(record)
	assert.Equal(t, err, nil)
	assert.Equal(t, record.Id, 1)

	record, err = store.Get(1)
	assert.Equal(t, err, nil)
	assert.Equal(t, record.Id, 1)
	assert.Equal(t, record.Value, []byte("foo"))
	assert.Equal(t, record.Queue, "testqueue")

	err = store.Remove(1)
	assert.Equal(t, err, nil)

	record, err = store.Get(1)
	assert.Equal(t, err, nil)
	assert.T(t, record == nil)
}
