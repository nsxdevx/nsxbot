package types

type EventType = string

const (
	POST_TYPE_MESSAGE    = "message"
	POST_TYPE_NOTICE     = "notice"
	POST_TYPE_REQUEST    = "request"
	POST_TYPE_META_ENEVT = "meta_event"
)

type Event struct {
	Types   []EventType
	Time    int64
	SelfID  int64
	RawData []byte
	Replyer Replyer
}

type Eventer interface {
	Type() EventType
}
