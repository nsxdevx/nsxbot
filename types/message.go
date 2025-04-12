package types

import (
	"encoding/json"
	"errors"
	"strings"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Text struct {
	Text string `json:"text"`
}

func (t Text) Cmd(prefix string) (string, error) {
	trimmed := strings.TrimSpace(t.Text)
	if !strings.HasPrefix(trimmed, prefix) {
		return "", errors.New("not a command")
	}
	parts := strings.Fields(strings.TrimLeft(trimmed, prefix))
	if len(parts) < 2 {
		return "", errors.New("not enough parts")
	}
	return parts[0], nil
}

func (t Text) CmdIndex(prefix string, index int) (string, error) {
	if _, err := t.Cmd(prefix); err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(t.Text)
	parts := strings.Fields(trimmed)
	if len(parts)-1 < index {
		return "", errors.New("index out of range")
	}
	return parts[index+1], nil
}

func (t Text) CmdKey(key string) (string, error) {
	trimmed := strings.TrimSpace(t.Text)
	parts := strings.Fields(trimmed)
	if len(parts) < 3 {
		return "", errors.New("not enough parts")
	}
	for i := 1; i+1 < len(parts); i = i + 2 {
		if strings.EqualFold(strings.ToLower(parts[i]), strings.ToLower(key)) {
			return parts[i+1], nil
		}
	}
	return "", errors.New("key not found")
}

type Face struct {
	Id string `json:"id"`
}

type At struct {
	QQ string `json:"qq"`
}

type Image struct {
	Name       string `json:"name"`
	Summary    string `json:"summary"`
	File       string `json:"file"` // marketface
	SubType    string `json:"subtype"`
	FileID     string `json:"file_id"`
	Url        string `json:"url"`
	Path       string `json:"path"`
	FileSize   string `json:"file_size"`
	FileUnique string `json:"file_unique"`
}
