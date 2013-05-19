package kew

import (
	"time"
)

var Queues = map[string]*Queue{}

const NilDuration = time.Duration(-1)

type Queue struct {
	Name  string
	Items chan *Item
}

func NewQueue(name string) *Queue {
	q, ok := Queues[name]
	if !ok {
		q = &Queue{name, make(chan *Item)}
		Queues[name] = q
	}
	return q
}

func (q *Queue) Enqueue(item *Item) {
	item.dequeued = false
	item.queue = q.Name
	q.Items <- item
}

func (q *Queue) Dequeue(wait time.Duration, timeout time.Duration) (*Item, bool) {
	if wait != NilDuration {
		// Blocking
		expired := time.After(wait)
		select {
		case <-expired:
			return nil, false
		case item := <-q.Items:
			item.dequeued = true
			go q.SetTimeout(item, timeout)
			return item, true
		}
	} else {
		// Nonblocking
		select {
		case item := <-q.Items:
			item.dequeued = true
			go q.SetTimeout(item, timeout)
			return item, true
		default:
			return nil, false
		}
	}

	return nil, false
}

func (q *Queue) SetTimeout(item *Item, timeout time.Duration) {
	if timeout == NilDuration {
		return
	}

	expired := time.After(timeout)

	select {
	case <-expired:
		q.Enqueue(item)
	case <-item.complete:
		return
	}
}
