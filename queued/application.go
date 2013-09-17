package queued

import (
	"sync"
	"time"
)

type Info struct {
	record   *Record
	dequeued bool
}

type Application struct {
	store  Store
	queues map[string]*Queue
	items  map[int]*Item
	qmutex sync.Mutex
	imutex sync.RWMutex
}

func NewApplication(store Store) *Application {
	app := &Application{
		store:  store,
		queues: make(map[string]*Queue),
		items:  make(map[int]*Item),
	}

	it := store.Iterator()
	record, ok := it.NextRecord()

	for ok {
		queue := app.GetQueue(record.Queue)
		item := queue.Enqueue(record.id)
		app.items[item.value] = item

		record, ok = it.NextRecord()
	}

	return app
}

func (a *Application) GetQueue(name string) *Queue {
	a.qmutex.Lock()
	defer a.qmutex.Unlock()

	queue, ok := a.queues[name]

	if !ok {
		queue = NewQueue()
		a.queues[name] = queue
	}

	return queue
}

func (a *Application) GetItem(id int) (*Item, bool) {
	a.imutex.RLock()
	defer a.imutex.RUnlock()

	item, ok := a.items[id]
	return item, ok
}

func (a *Application) PutItem(item *Item) {
	a.imutex.Lock()
	defer a.imutex.Unlock()

	a.items[item.value] = item
}

func (a *Application) RemoveItem(id int) {
	a.imutex.Lock()
	defer a.imutex.Unlock()

	delete(a.items, id)
}

func (a *Application) Enqueue(name string, value []byte, mime string) (*Record, error) {
	queue := a.GetQueue(name)
	record := NewRecord(value, name)
	record.Mime = mime

	err := a.store.Put(record)
	if err != nil {
		return nil, err
	}

	item := queue.Enqueue(record.id)
	a.PutItem(item)

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
	item, ok := a.GetItem(id)

	if !ok || !item.dequeued {
		return false, nil
	}

	err := a.store.Remove(id)
	if err != nil {
		return false, err
	}

	item.Complete()
	a.RemoveItem(id)

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

	item, ok := a.GetItem(id)
	info := &Info{record, ok && item.dequeued}

	return info, nil
}

func (a *Application) Stats(name string) map[string]int {
	queue := a.GetQueue(name)
	return queue.Stats()
}
