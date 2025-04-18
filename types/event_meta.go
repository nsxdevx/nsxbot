package types

type LifeMeta struct {
	SubType string `json:"sub_type"` //enable disable connect
}

func (e LifeMeta) Type() EventType {
	return "meta_event:lifecycle"
}

type HeartMeta struct {
	Status   Status `json:"status"`
	Interval int64  `json:"interval"`
}

func (e HeartMeta) Type() EventType {
	return "meta_event:heartbeat"
}
