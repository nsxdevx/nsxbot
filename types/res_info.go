package types

type LoginInfo struct {
	UserID   int64  `json:"user_id"`
	NickName string `json:"nickname"`
}

type Status struct {
	Online bool `json:"online"`
	Good   bool `json:"good"`
}
