package queued

import (
	"time"
)

const NilDuration = time.Duration(-1)

type Queue struct {
	items   []*Item
	dequeue chan *Reader
	enqueue chan *Item
	waiting chan *Item
	running bool
}

func NewQueue() *Queue {
	q := &Queue{
		items:   []*Item{},
		dequeue: make(chan *Reader),
		enqueue: make(chan *Item),
		waiting: make(chan *Item),
		running: true,
	}

	go q.run()

	return q
}

func (q *Queue) Enqueue(value int) *Item {
	item := NewItem(value)
	q.EnqueueItem(item)
	return item
}

func (q *Queue) EnqueueItem(item *Item) {
	item.dequeued = false

	select {
	case q.waiting <- item:
	default:
		q.enqueue <- item
	}
}

func (q *Queue) Dequeue(wait time.Duration, timeout time.Duration) *Item {
	reader := NewReader(wait, timeout, q.expire)

	go func() {
		q.dequeue <- reader
	}()

	return reader.Receive()
}

func (q *Queue) Stop() {
	q.running = false
}

func (q *Queue) run() {
	for q.running {
		select {
		case item := <-q.enqueue:
			q.items = append(q.items, item)
		case reader := <-q.dequeue:
			if item := q.shift(); item != nil {
				reader.Send(item)
			} else {
				reader.Wait(q.waiting)
			}
		}
	}
}

func (q *Queue) expire(item *Item) {
	q.EnqueueItem(item)
}

func (q *Queue) shift() *Item {
	if len(q.items) > 0 {
		item := q.items[0]
		q.items = q.items[1:]
		return item
	} else {
		return nil
	}
}
