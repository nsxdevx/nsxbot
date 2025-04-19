package types

type EventFriendReq struct {
	UserId  int64  `json:"user_id"`
	Comment string `json:"comment"`
	Flag    string `json:"flag"`
}

func (fr EventFriendReq) Type() EventType {
	return "request:friend"
}

func (fr *EventFriendReq) Reply(replyer Replayer, approve bool, remark string) error {
	if replyer == nil {
		return ErrNoAvailable
	}
	data := struct {
		Approve bool   `json:"approve"`
		Remark  string `json:"remark"`
	}{Approve: approve, Remark: remark}
	return replyer.Reply(data)
}

type EventGroupReq struct {
	SubType string `json:"sub_type"` // add invite
	GroupId int64  `json:"group_id"`
	UserId  int64  `json:"user_id"`
	Comment string `json:"comment"`
	Flag    string `json:"flag"`
}

func (gr EventGroupReq) Type() EventType {
	return "request:group"
}

func (gr *EventGroupReq) Reply(replyer Replayer, approve bool, reason string) error {
	if replyer == nil {
		return ErrNoAvailable
	}
	data := struct {
		Approve bool   `json:"approve"`
		Reason  string `json:"reason"`
	}{Approve: approve, Reason: reason}
	return replyer.Reply(data)
}
