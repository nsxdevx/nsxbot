package types

type SendPrivateMsgReq struct {
	UserId  int64     `json:"user_id"`
	Message []Message `json:"message"`
}

type SendGrMsgReq struct {
	GroupId int64     `json:"group_id"`
	Message []Message `json:"message"`
}

type GetStrangerInfo struct {
	UserId  int64 `json:"user_id"`
	NoCache bool  `json:"no_cache"`
}

type GetMsgReq struct {
	MessageId int32 `json:"message_id"`
}

type DelMsgReq struct {
	MessageId int32 `json:"message_id"`
}
