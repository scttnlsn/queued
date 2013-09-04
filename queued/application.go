package queued

import (
	"time"
)

type Info struct {
	value    []byte
	dequeued bool
}

type Application struct {
	store  *Store
	queues map[string]*Queue
	items  map[int]*Item
}

func NewApplication(store *Store) *Application {
	app := &Application{
		store:  store,
		queues: make(map[string]*Queue),
		items:  make(map[int]*Item),
	}

	it := store.Iterator()

	for it.Valid() {
		record := it.Record()
		queue := app.GetQueue(record.Queue)
		item := queue.Enqueue(record.id)

		app.items[item.value] = item

		it.Next()
	}

	return app
}

func (a *Application) GetQueue(name string) *Queue {
	queue, ok := a.queues[name]

	if !ok {
		queue = NewQueue()
		a.queues[name] = queue
	}

	return queue
}

func (a *Application) Enqueue(name string, value []byte) (*Record, error) {
	queue := a.GetQueue(name)
	record := NewRecord(value, name)

	err := a.store.Put(record)
	if err != nil {
		return nil, err
	}

	item := queue.Enqueue(record.id)
	a.items[item.value] = item

	return record, nil
}

func (a *Application) Dequeue(name string, wait time.Duration, timeout time.Duration) (*Record, error) {
	queue := a.GetQueue(name)
	item := queue.Dequeue(wait, timeout)
	if item == nil {
		return nil, nil
	}

	record, err := a.store.Get(item.value)
	if err != nil {
		return nil, err
	}

	if !item.dequeued {
		a.Complete(name, item.value)
	}

	return record, nil
}

func (a *Application) Complete(name string, id int) (bool, error) {
	item, ok := a.items[id]
	if !ok || !item.dequeued {
		return false, nil
	}

	err := a.store.Remove(id)
	if err != nil {
		return false, err
	}

	item.Complete()
	delete(a.items, id)

	return true, nil
}

func (a *Application) Info(name string, id int) (*Info, error) {
	record, err := a.store.Get(id)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}

	if record.Queue != name {
		return nil, nil
	}

	item, ok := a.items[id]
	info := &Info{record.Value, ok && item.dequeued}

	return info, nil
}
