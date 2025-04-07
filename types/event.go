package types

import (
	"context"
	"encoding/json"
	"errors"
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
	Messages    []Message `json:"message"`
	RawMessage  string    `json:"raw_message"`
	Font        int32     `json:"font"`
	Sender      Sender    `json:"sender"`
}

var (
	ErrNotFound      = errors.New("not found")
	ErrTypeAssertion = errors.New("type assertion failed")
)

func (em *EventMessage) TextFirst() (*Text, error) {
	return first[Text]("text", em.Messages)
}

func (em *EventMessage) Texts() ([]Text, int) {
	return all[Text]("text", em.Messages)
}

func (em *EventMessage) Faces() ([]Face, int) {
	return all[Face]("face", em.Messages)
}

func (em *EventMessage) FaceFirst() (*Face, error) {
	return first[Face]("face", em.Messages)
}

func first[T any](msgType string, msg []Message) (*T, error) {
	for _, msg := range msg {
		if msg.Type == msgType {
			var data T
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				return nil, err
			}
			return &data, nil
		}
	}
	return nil, ErrNotFound
}

func all[T any](msgType string, msg []Message) ([]T, int) {
	var data []T
	for _, msg := range msg {
		if msg.Type == msgType {
			var d T
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				continue
			}
			data = append(data, d)
		}
	}
	return data, len(data)
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

type EventMeta struct {
}
