package kew

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestNewItem(t *testing.T) {
	item := NewItem("foo")
	assert.Equal(t, item.value, "foo")
	assert.Equal(t, item.dequeued, false)
	assert.Equal(t, item.id, 0)
}

func TestCompleteItem(t *testing.T) {
	foo := NewItem("foo")
	foo.dequeued = true
	go foo.Complete()
	<-foo.complete

	bar := NewItem("bar")
	assert.Equal(t, bar.Complete(), false)
}
