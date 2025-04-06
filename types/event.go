package types

import (
	"context"
	"encoding/json"
	"net/http"
)

type PostType = string

const (
	POST_TYPE_MESSAGE    = "message"
	POST_TYPE_NOTICE     = "notice"
	POST_TYPE_REQUEST    = "request"
	POST_TYPE_META_ENEVT = "meta_event"
)

type Event struct {
	PostType PostType
	Time     int64
	SelfID   int64
	RawData  []byte
	Replyer  *Replyer
}

type Replyer struct {
	Ctx    context.Context
	Writer http.ResponseWriter
}

func (r *Replyer) Reply(text string) error {
	if r.Ctx.Err() != nil {
		return r.Ctx.Err()
	}
	body, err := json.Marshal(struct {
		Reply string `json:"reply"`
	}{Reply: text})
	if err != nil {
		return err
	}
	_, err = r.Writer.Write(body)
	return err
}

type EventMessage struct {
	MessageType string    `json:"message_type"`
	SubType     string    `json:"sub_type"`
	MessageID   int32     `json:"message_id"`
	UserID      int64     `json:"user_id"`
	GroupID     int64     `json:"group_id"`
	Message     []Message `json:"message"`
	RawMessage  string    `json:"raw_message"`
	Font        int32     `json:"font"`
	Sender      Sender    `json:"sender"`
}

type Message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type Sender struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int32  `json:"age"`
}

type EventNotice struct {
}

type EventRequest struct {
}

type EventMetaEvent struct {
}
