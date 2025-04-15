package driver

type Action = string

const (
	ACTION_SEND_PRIVATE_MSG  = "send_private_msg"
	ACTION_SEND_GROUP_MSG    = "send_group_msg"
	ACTION_GET_MSG           = "get_msg"
	ACTION_GET_LOGIN_INFO    = "get_login_info"
	ACTION_GET_STRANGER_INFO = "get_stranger_info"
	Action_GET_STATUS        = "get_status"
)
