package queued

import (
	"github.com/bmizerany/assert"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	q := NewQueue()

	q.Enqueue(123)
	q.Enqueue(456)

	one := q.Dequeue(NilDuration, NilDuration)
	assert.Equal(t, one.value, 123)

	two := q.Dequeue(NilDuration, NilDuration)
	assert.Equal(t, two.value, 456)
}

func TestDequeueWait(t *testing.T) {
	q := NewQueue()

	wait := time.Millisecond

	go func() {
		time.Sleep(wait)
		q.Enqueue(123)
	}()

	one := q.Dequeue(NilDuration, NilDuration)
	assert.T(t, one == nil)

	two := q.Dequeue(time.Second, NilDuration)
	assert.T(t, two != nil)
	assert.Equal(t, two.value, 123)
}

func TestDequeueTimeout(t *testing.T) {
	q := NewQueue()

	timeout := time.Millisecond

	q.Enqueue(123)

	one := q.Dequeue(NilDuration, timeout)
	assert.T(t, one != nil)

	time.Sleep(timeout * 2)

	two := q.Dequeue(NilDuration, timeout)
	assert.T(t, two != nil)

	two.Complete()
	time.Sleep(timeout)

	three := q.Dequeue(NilDuration, NilDuration)
	assert.T(t, three == nil)
}

func TestStats(t *testing.T) {
	q := NewQueue()

	q.Enqueue(123)
	q.Enqueue(456)

	assert.Equal(t, q.Stats()["enqueued"], 2)
	assert.Equal(t, q.Stats()["dequeued"], 0)
	assert.Equal(t, q.Stats()["depth"], 2)

	one := q.Dequeue(NilDuration, NilDuration)
	assert.Equal(t, one.value, 123)

	assert.Equal(t, q.Stats()["enqueued"], 2)
	assert.Equal(t, q.Stats()["dequeued"], 1)
	assert.Equal(t, q.Stats()["depth"], 1)

	two := q.Dequeue(NilDuration, NilDuration)
	assert.Equal(t, two.value, 456)

	assert.Equal(t, q.Stats()["enqueued"], 2)
	assert.Equal(t, q.Stats()["dequeued"], 2)
	assert.Equal(t, q.Stats()["depth"], 0)

	q.Enqueue(789)
	q.Dequeue(NilDuration, time.Millisecond)
	time.Sleep(time.Millisecond * 2)

	assert.Equal(t, q.Stats()["enqueued"], 4)
	assert.Equal(t, q.Stats()["dequeued"], 3)
	assert.Equal(t, q.Stats()["depth"], 1)
	assert.Equal(t, q.Stats()["timeouts"], 1)
}
