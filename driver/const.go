package driver

type Action = string

const (
	ACTION_SEND_PRIVATE_MSG = "send_private_msg"
	ACTION_SEND_GROUP_MSG   = "send_group_msg"
	ACTION_GET_LOGIN_INFO   = "get_login_info"
	Action_GET_STATUS       = "get_status"
)
