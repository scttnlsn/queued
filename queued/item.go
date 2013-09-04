package queued

type Item struct {
	value    int
	dequeued bool
	complete chan bool
}

func NewItem(value int) *Item {
	return &Item{
		value:    value,
		dequeued: false,
		complete: make(chan bool),
	}
}

func (i *Item) Complete() {
	i.complete <- true
}
