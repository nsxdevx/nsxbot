package types

import (
	"encoding/json"
	"errors"
	"fmt"
)

type EventMsg interface {
	Eventer
	TextFirst() (*Text, error)
	Texts() ([]Text, int)
	FaceFirst() (*Face, error)
	Faces() ([]Face, int)
	AtFirst() (*At, error)
	Ats() ([]At, int)
	ImageFirst() (*Image, error)
	Images() ([]Image, int)
}

type BaseMessage struct {
	SubType    string    `json:"sub_type"`
	MessageId  int       `json:"message_id"`
	UserId     int64     `json:"user_id"`
	Messages   []Message `json:"message"`
	RawMessage string    `json:"raw_message"`
	Font       int       `json:"font"`
	Sender     Sender    `json:"sender"`
}

type Sender struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int    `json:"age"`
}

func (em *BaseMessage) Reply(replyer Replyer, text string) error {
	if replyer == nil {
		return ErrNoAvailable
	}
	data := struct {
		Reply string `json:"reply"`
	}{Reply: text}
	return replyer.Reply(data)
}

func (em *BaseMessage) Id() int {
	return em.MessageId
}

func (em *BaseMessage) TextFirst() (*Text, error) {
	return first[Text]("text", em.Messages)
}

func (em *BaseMessage) Texts() ([]Text, int) {
	return all[Text]("text", em.Messages)
}

func (em *BaseMessage) Faces() ([]Face, int) {
	return all[Face]("face", em.Messages)
}

func (em *BaseMessage) FaceFirst() (*Face, error) {
	return first[Face]("face", em.Messages)
}

func (em *BaseMessage) AtFirst() (*At, error) {
	return first[At]("at", em.Messages)
}

func (em *BaseMessage) Ats() ([]At, int) {
	return all[At]("at", em.Messages)
}

func (em *BaseMessage) ReplyFirst() (*Reply, error) {
	return first[Reply]("reply", em.Messages)
}

func (em *BaseMessage) ImageFirst() (*Image, error) {
	return first[Image]("image", em.Messages)
}

func (em *BaseMessage) Images() ([]Image, int) {
	return all[Image]("image", em.Messages)
}

var (
	ErrNotFound = errors.New("not found")
)

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

type EventPvtMsg struct {
	*BaseMessage
}

func (e EventPvtMsg) Type() EventType {
	return "message:private"
}

func (e EventPvtMsg) SessionKey() string {
	return e.Type() + ":" + fmt.Sprint(e.UserId)
}

type EventGrMsg struct {
	*BaseMessage
	GroupId   int64      `json:"group_id"`
	Anonymous *Anonymous `json:"anonymous"`
}

type Anonymous struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Flag string `json:"flag"`
}

func (e EventGrMsg) Type() EventType {
	return "message:group"
}

func (e EventGrMsg) SessionKey() string {
	return e.Type() + ":" + fmt.Sprint(e.GroupId) + ":" + fmt.Sprint(e.UserId)
}

type EventAllMsg struct {
	*BaseMessage
	GroupId   int64      `json:"group_id"`
	Anonymous *Anonymous `json:"anonymous"`
}

func (em EventAllMsg) Type() EventType {
	return "message"
}

func (em EventAllMsg) SessionKey() string {
	return em.Type() + ":" + fmt.Sprint(em.GroupId) + ":" + fmt.Sprint(em.UserId)
}
