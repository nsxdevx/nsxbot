package types

type LoginInfo struct {
	UserId   int64  `json:"user_id"`
	NickName string `json:"nickname"`
}

type StrangerInfo struct {
	UserId   int64  `json:"user_id"`
	NickName string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int32  `json:"age"`
}

type Status struct {
	Online bool `json:"online"`
	Good   bool `json:"good"`
}
