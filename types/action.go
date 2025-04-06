package types

type Action = string

const (
	ACTION_GET_LOGIN_INFO = "get_login_info"
)

type Request[T any] struct {
	Echo   int64  `json:"echo"`
	Action Action `json:"action"`
	Params T      `json:"params,omitempty"`
}

type Response[T any] struct {
	Status  string `json:"status"`
	RetCode int    `json:"retCode"`
	Data    T      `json:"data,omitempty"`
}
