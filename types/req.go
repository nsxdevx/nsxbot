package types

type SendPrivateMsgReq struct {
	UserId  int64     `json:"user_id"`
	Message []Message `json:"message"`
}

type SendGrMsgReq struct {
	GroupId int64     `json:"group_id"`
	Message []Message `json:"message"`
}
