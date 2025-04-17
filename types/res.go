package types

type SendMsgRes struct {
	MessageId int `json:"message_id"`
}

type GetMsgRes struct {
	Time        int       `json:"time"`
	MessageType string    `json:"message_type"`
	MessageId   int       `json:"message_id"`
	RealId      int       `json:"real_id"`
	Sender      Sender    `json:"sender"`
	Message     []Message `json:"message"`
}
