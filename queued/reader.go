package queued

import (
	"time"
)

type Reader struct {
	wait    time.Duration
	timeout time.Duration
	expire  func(*Item)
	ch      chan *Item
}

func NewReader(wait time.Duration, timeout time.Duration, expire func(*Item)) *Reader {
	return &Reader{
		wait:    wait,
		timeout: timeout,
		expire:  expire,
		ch:      make(chan *Item),
	}
}

func (r *Reader) Send(item *Item) {
	r.ch <- item
}

func (r *Reader) Receive() *Item {
	item := <-r.ch
	r.Timeout(item)
	return item
}

func (r *Reader) Wait(ch chan *Item) {
	if r.wait != NilDuration {
		go func() {
			select {
			case <-time.After(r.wait):
				r.Send(nil)
			case item := <-ch:
				r.Send(item)
			}
		}()
	} else {
		r.Send(nil)
	}
}

func (r *Reader) Timeout(item *Item) {
	if r.timeout != NilDuration {
		item.dequeued = true

		go func() {
			select {
			case <-time.After(r.timeout):
				r.expire(item)
			case <-item.complete:
				return
			}
		}()
	}
}
