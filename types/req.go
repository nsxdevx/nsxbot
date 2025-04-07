package types

type SendPrivateMsgReq struct {
	UserId  int64     `json:"user_id"`
	Message []Message `json:"message"`
}
