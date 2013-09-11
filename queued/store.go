package queued

type Iterator interface {
	NextRecord() (*Record, bool)
}

type Store interface {
	Get(id int) (*Record, error)
	Put(record *Record) error
	Remove(id int) error
	Iterator() Iterator
	Drop()
}
