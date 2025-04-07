package types

import "encoding/json"

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
