package driver

import (
	"context"
	"fmt"

	"github.com/nsxdevx/nsxbot/event"
	"github.com/nsxdevx/nsxbot/schema"
	"github.com/nsxdevx/nsxbot/types"
	"github.com/tidwall/gjson"
)

type Driver interface {
	EmitterMux
	Listener
}

type Listener interface {
	Listen(ctx context.Context, eventChan chan<- event.Event) error
}

type EmitterMux interface {
	GetEmitter(selfId int64) (Emitter, error)
	AddEmitter(selfId int64, emitter Emitter)
}

type Emitter interface {
	SendPvtMsg(ctx context.Context, userId int64, msg schema.MessageChain) (*types.SendMsgRes, error)
	SendGrMsg(ctx context.Context, groupId int64, msg schema.MessageChain) (*types.SendMsgRes, error)
	GetMsg(ctx context.Context, msgId int) (*types.GetMsgRes, error)
	DelMsg(ctx context.Context, msgId int) error
	GetLoginInfo(ctx context.Context) (*types.LoginInfo, error)
	GetStrangerInfo(ctx context.Context, userId int64, noCache bool) (*types.StrangerInfo, error)
	GetStatus(ctx context.Context) (*types.Status, error)
	GetVersionInfo(ctx context.Context) (*types.VersionInfo, error)
	GetSelfId(ctx context.Context) (int64, error)
	SetFriendAddRequest(ctx context.Context, flag string, approve bool, remark string) error
	SetGroupAddRequest(ctx context.Context, flag string, approve bool, reason string) error
	SetGroupSpecialTitle(ctx context.Context, groupId int64, userId int64, specialTitle string, duration int) error
	Raw(ctx context.Context, action Action, params any) ([]byte, error)
}

type Request[T any] struct {
	Echo   string `json:"echo"`
	Action Action `json:"action"`
	Params T      `json:"params,omitempty"`
}

type Response[T any] struct {
	Status  string `json:"status"`
	RetCode int    `json:"retCode"`
	Data    T      `json:"data,omitempty"`
	Echo    string `json:"echo"`
}

func contentToEvent(content []byte) (event.Event, error) {
	strContent := string(content)
	postType := gjson.Get(strContent, "post_type")
	if !postType.Exists() {
		return event.Event{}, fmt.Errorf("invalid event, post_type: %v", postType.Exists())
	}

	Type := gjson.Get(strContent, postType.String()+"_type")
	if !Type.Exists() {
		return event.Event{}, fmt.Errorf("invalid event, %s_type: %v", postType.String(), Type.Exists())
	}

	time := gjson.Get(strContent, "time")
	selfId := gjson.Get(strContent, "self_id")
	if !time.Exists() || !selfId.Exists() {
		return event.Event{}, fmt.Errorf("invalid event, post_type: %v, time: %v, self_id: %v", postType.Exists(), time.Exists(), selfId.Exists())
	}

	return event.Event{
		Types:   []string{postType.String(), postType.String() + ":" + Type.String()},
		RawData: content,
		SelfId:  selfId.Int(),
		Time:    time.Int(),
	}, nil
}
