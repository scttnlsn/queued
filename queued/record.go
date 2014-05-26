package queued

type Record struct {
	Id    int
	Value []byte
	Mime  string
	Queue string
}

func NewRecord(value []byte, queue string) *Record {
	return &Record{
		Id:    0,
		Value: value,
		Mime:  "",
		Queue: queue,
	}
}

func (r *Record) ContentType() string {
	if r.Mime == "" {
		return "application/octet-stream"
	} else {
		return r.Mime
	}
}
