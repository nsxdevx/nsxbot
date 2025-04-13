package types

type SendMsgRes struct {
	MessageId int32 `json:"message_id"`
}

type GetMsgRes struct {
	Time        int32     `json:"time"`
	MessageType string    `json:"message_type"`
	MessageId   int32     `json:"message_id"`
	RealId      int32     `json:"real_id"`
	Sender      Sender    `json:"sender"`
	Message     []Message `json:"message"`
}
