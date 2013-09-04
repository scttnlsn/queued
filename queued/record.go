package queued

type Record struct {
	id    int
	Value []byte
	Queue string
}

func NewRecord(value []byte, queue string) *Record {
	return &Record{
		id:    0,
		Value: value,
		Queue: queue,
	}
}
