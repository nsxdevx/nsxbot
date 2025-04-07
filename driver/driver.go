package driver

import (
	"context"
	"fmt"

	"github.com/atopos31/nsxbot/types"
	"github.com/tidwall/gjson"
)

type Driver interface {
	Emitter
	Listener
}

type Listener interface {
	Listen(ctx context.Context, eventChan chan<- types.Event) error
}

type Emitter interface {
	SendPrivateMsg(ctx context.Context, userId int64, msg types.MeaasgeChain) (*types.SendMsgRes, error)
	GetLoginInfo(ctx context.Context) (*types.LoginInfo, error)
	Raw(ctx context.Context, action types.Action, params any) ([]byte, error)
}

type Request[T any] struct {
	Echo   int64        `json:"echo"`
	Action types.Action `json:"action"`
	Params T            `json:"params,omitempty"`
}

type Response[T any] struct {
	Status  string `json:"status"`
	RetCode int    `json:"retCode"`
	Data    T      `json:"data,omitempty"`
}

func contentToEvent(content []byte) (types.Event, error) {
	strContent := string(content)
	postType := gjson.Get(strContent, "post_type")
	time := gjson.Get(strContent, "time")
	selfID := gjson.Get(strContent, "self_id")
	if !postType.Exists() || !time.Exists() || !selfID.Exists() {
		return types.Event{}, fmt.Errorf("invalid event, post_type: %v, time: %v, self_id: %v", postType.Exists(), time.Exists(), selfID.Exists())
	}
	return types.Event{
		PostType: postType.String(),
		RawData:  content,
		SelfID:   selfID.Int(),
		Time:     time.Int(),
	}, nil
}
