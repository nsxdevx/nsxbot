package types

import "encoding/json"

type MeaasgeChain []Message

func (m MeaasgeChain) append(msg Message) MeaasgeChain {
	return append(m, msg)
}

func (m MeaasgeChain) Text(text string) MeaasgeChain {
	data, err := json.Marshal(Text{
		Text: text,
	})
	if err != nil {
		panic(err)
	}
	return m.append(Message{
		Type: "text",
		Data: data,
	})
}

func (m MeaasgeChain) Face(id string) MeaasgeChain {
	data, err := json.Marshal(Face{
		Id: id,
	})
	if err != nil {
		panic(err)
	}
	return m.append(Message{
		Type: "face",
		Data: data,
	})
}

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Text struct {
	Text string `json:"text"`
}

type Face struct {
	Id string `json:"id"`
}

type Image struct {
	Name       string `json:"name"`
	Summary    string `json:"summary"`
	File       string `json:"file"` // marketface
	SubType    string `json:"subtype"`
	FileID     string `json:"file_id"`
	Url        string `json:"url"`
	Path       string `json:"path"`
	FileSize   int64  `json:"file_size"`
	FileUnique string `json:"file_unique"`
}
