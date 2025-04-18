package types

import "encoding/json"

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
	data, err := json.Marshal(struct {
		Approve bool   `json:"approve"`
		Remark  string `json:"remark"`
	}{Approve: approve, Remark: remark})
	if err != nil {
		return err
	}
	return replyer.Reply(data)
}
