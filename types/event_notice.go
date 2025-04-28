package types

type EventNotify struct {
	SubType  string `json:"sub_type"`
	TargetId int64  `json:"target_id"`
	UserId   int64  `json:"user_id"`
	GroupId  int64  `json:"group_id"`
}

func (en EventNotify) Type() EventType {
	return "notice:notify"
}

type EventGrRecall struct {
	GroupId    int64 `json:"group_id"`
	UserId     int64 `json:"user_id"`
	OperatorId int64 `json:"operator_id"`
	MessageId  int64 `json:"message_id"`
}

func (en EventGrRecall) Type() EventType {
	return "notice:group_recall"
}

type EventPvtRecall struct {
	UserId    int64 `json:"user_id"`
	MessageId int64 `json:"message_id"`
}

func (en EventPvtRecall) Type() EventType {
	return "notice:friend_recall"
}

type EventGrDecrease struct {
	SubType    string `json:"sub_type"` // leave/kick/kick_me
	GroupId    int64  `json:"group_id"`
	UserId     int64  `json:"user_id"`
	OperatorId int64  `json:"operator_id"`
}

func (en EventGrDecrease) Type() EventType {
	return "notice:group_decrease"
}

type EventGrIncrease struct {
	SubType    string `json:"sub_type"` // approve/invite
	GroupId    int64  `json:"group_id"`
	UserId     int64  `json:"user_id"`
	OperatorId int64  `json:"operator_id"`
}

func (en EventGrIncrease) Type() EventType {
	return "notice:group_increase"
}

type EventAdmin struct {
	SubType string `json:"sub_type"` // set/unset
	GroupId int64  `json:"group_id"`
	UserId  int64  `json:"user_id"`
}

func (en EventAdmin) Type() EventType {
	return "notice:group_admin"
}

type EventGrFile struct {
	GroupId int64 `json:"group_id"`
	UserId  int64 `json:"user_id"`
	File    GrFile  `json:"file"`
}

type GrFile struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Busid int64  `json:"busid"`
}

func (en EventGrFile) Type() EventType {
	return "notice:group_upload"
}

type EventGrBan struct {
	SubType    string `json:"sub_type"` // ban/lift_ban
	GroupId    int64  `json:"group_id"`
	UserId     int64  `json:"user_id"`
	Duration   int64  `json:"duration"` // s
	OperatorId int64  `json:"operator_id"`
}

type EventPvtAdd struct {
	UserId int64 `json:"user_id"`
}

func (en EventPvtAdd) Type() EventType {
	return "notice:friend_add"
}
