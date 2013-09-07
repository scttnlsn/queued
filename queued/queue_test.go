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
