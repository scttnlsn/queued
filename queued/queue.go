package queued

import (
	"sync"
	"time"
)

const NilDuration = time.Duration(-1)

type Queue struct {
	items   []*Item
	waiting chan *Item
	depth   int
	mutex   sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		items:   []*Item{},
		waiting: make(chan *Item),
	}
}

func (q *Queue) Enqueue(value int) *Item {
	item := NewItem(value)
	q.EnqueueItem(item)
	return item
}

func (q *Queue) EnqueueItem(item *Item) {
	select {
	case q.waiting <- item:
	default:
		q.append(item)
	}
}

func (q *Queue) Dequeue(wait time.Duration, timeout time.Duration) *Item {
	if item := q.shift(); item != nil {
		q.timeout(item, timeout)
		return item
	} else if wait != NilDuration {
		select {
		case <-time.After(wait):
			return nil
		case item := <-q.waiting:
			q.timeout(item, timeout)
			return item
		}
	} else {
		return nil
	}
}

func (q *Queue) shift() *Item {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.items) > 0 {
		item := q.items[0]
		q.items = q.items[1:]
		q.depth -= 1
		return item
	} else {
		return nil
	}
}

func (q *Queue) append(item *Item) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.items = append(q.items, item)
	q.depth += 1
}

func (q *Queue) timeout(item *Item, timeout time.Duration) {
	if timeout != NilDuration {
		item.dequeued = true

		go func() {
			select {
			case <-time.After(timeout):
				item.dequeued = false
				q.EnqueueItem(item)
			case <-item.complete:
				item.dequeued = false
				return
			}
		}()
	}
}
