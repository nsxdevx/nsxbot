package filter

import (
	"errors"
	"testing"

	"github.com/nsxdevx/nsxbot/schema"
	"github.com/stretchr/testify/assert"
)

// MockEventMsg 是 EventMsg 的模拟实现。
type MockEventMsg struct {
	text string
	err  error
}

func (m MockEventMsg) TextFirst() (*schema.Text, error) {
	return &schema.Text{Text: m.text}, m.err
}

func (m MockEventMsg) AtFirst() (*schema.At, error) {
	return &schema.At{}, nil
}

func (m MockEventMsg) Ats() ([]schema.At, int) {
	return []schema.At{}, 0
}

func (m MockEventMsg) FaceFirst() (*schema.Face, error) {
	return &schema.Face{}, nil
}

func (m MockEventMsg) Faces() ([]schema.Face, int) {
	return []schema.Face{}, 0
}

func (m MockEventMsg) Texts() ([]schema.Text, int) {
	return []schema.Text{}, 0
}

func (m MockEventMsg) ImageFirst() (*schema.Image, error) {
	return &schema.Image{}, nil
}

func (m MockEventMsg) Images() ([]schema.Image, int) {
	return []schema.Image{}, 0
}

func (m MockEventMsg) Type() string {
	return "mock"
}

func TestOnCommand(t *testing.T) {
	tests := []struct {
		msg      MockEventMsg
		prefix   string
		commands []string
		expected bool
	}{
		{MockEventMsg{"  !hello ", nil}, "!", []string{"hello"}, true},
		{MockEventMsg{"  !hello ", nil}, "", []string{"hello"}, false},
		{MockEventMsg{"  !hellow ", nil}, "!", []string{"hello"}, false},
		{MockEventMsg{"  !hello ", nil}, "/", []string{"hello"}, false},
		{MockEventMsg{"  !hi", nil}, "!", []string{"hello"}, false},
		{MockEventMsg{"  !hi", nil}, "!", []string{"hello", "hi"}, true},
		{MockEventMsg{"  !", nil}, "!", []string{"hello"}, false},
		{MockEventMsg{"  ", nil}, "!", []string{"hello"}, false},
		{MockEventMsg{"", errors.New("error")}, "!", []string{"hello"}, false},
	}

	for _, test := range tests {
		filter := OnCommand[MockEventMsg](test.prefix, test.commands...)
		result := filter(test.msg)
		assert.Equal(t, test.expected, result)
	}
}
