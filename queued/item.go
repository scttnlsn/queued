package queued

type Item struct {
	id       int
	value    string
	queue    string
	dequeued bool
	complete chan bool
}

func NewItem(value string) *Item {
	complete := make(chan bool)
	item := &Item{value: value, complete: complete}
	return item
}

func (i *Item) Complete() bool {
	ok := false

	if i.dequeued {
		i.complete <- true
		ok = true
	}

	return ok
}
