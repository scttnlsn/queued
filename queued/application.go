package queued

import (
	"time"
)

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
		app.GetQueue(record.Queue).Enqueue(record.id)
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

	queue.Enqueue(record.id)

	return record, nil
}

func (a *Application) Dequeue(name string, wait time.Duration, timeout time.Duration) (*Record, error) {
	queue := a.GetQueue(name)
	item := queue.Dequeue(wait, timeout)
	if item == nil {
		return nil, nil
	}

	id := item.value

	record, err := a.store.Get(id)
	if err != nil {
		return nil, err
	}

	if item.dequeued {
		a.items[id] = item
	} else {
		err := a.store.Remove(id)
		if err != nil {
			return nil, err
		}
	}

	return record, nil
}

func (a *Application) Complete(name string, id int) (bool, error) {
	item, ok := a.items[id]
	if !ok {
		return false, nil
	}

	err := a.store.Remove(id)
	if err != nil {
		return false, err
	}

	delete(a.items, id)
	item.Complete()

	return true, nil
}

func (a *Application) Info(name string, id int) (map[string]interface{}, error) {
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

	info := map[string]interface{}{
		"id":       record.id,
		"queue":    record.Queue,
		"dequeued": ok && item.dequeued,
	}

	return info, nil
}
