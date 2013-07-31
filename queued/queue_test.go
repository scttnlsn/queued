package queued

import (
	"github.com/bmizerany/assert"
	"testing"
	"time"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue("q1")
	assert.Equal(t, Queues["q1"], q)
	assert.Equal(t, NewQueue("q1"), q)
}

func TestEnqueue(t *testing.T) {
	q := NewQueue("q2")
	item := NewItem("foo")

	go q.Enqueue(item)
	<-q.Items

	assert.Equal(t, item.dequeued, false)
	assert.Equal(t, item.queue, "q2")
}

func TestDequeueNonBlocking(t *testing.T) {
	q := NewQueue("q3")

	_, ok := q.Dequeue(NilDuration, NilDuration)
	assert.Equal(t, ok, false)

	go q.Enqueue(NewItem("foo"))
	time.Sleep(time.Millisecond)

	item, ok := q.Dequeue(NilDuration, NilDuration)
	assert.Equal(t, item.value, "foo")
	assert.Equal(t, ok, true)
}

func TestDequeueBlocking(t *testing.T) {
	q := NewQueue("q4")
	go q.Enqueue(NewItem("foo"))

	item, ok := q.Dequeue(time.Hour, NilDuration)
	assert.Equal(t, item.value, "foo")
	assert.Equal(t, ok, true)
}

func TestDequeueTimeout(t *testing.T) {
	q := NewQueue("q5")
	item := NewItem("foo")

	go q.Enqueue(item)
	q.Dequeue(time.Hour, time.Millisecond)

	assert.Equal(t, item.dequeued, true)
	time.Sleep(time.Millisecond * 2)
	assert.Equal(t, item.dequeued, false)
}
