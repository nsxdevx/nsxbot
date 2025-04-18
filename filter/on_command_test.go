package filter

import (
	"errors"
	"testing"

	"github.com/atopos31/nsxbot/types"
	"github.com/stretchr/testify/assert"
)

// MockEventMsg 是 EventMsg 的模拟实现。
type MockEventMsg struct {
	text string
	err  error
}

func (m MockEventMsg) TextFirst() (*types.Text, error) {
	return &types.Text{Text: m.text}, m.err
}

func (m MockEventMsg) AtFirst() (*types.At, error) {
	return &types.At{}, nil
}

func (m MockEventMsg) Ats() ([]types.At, int) {
	return []types.At{}, 0
}

func (m MockEventMsg) FaceFirst() (*types.Face, error) {
	return &types.Face{}, nil
}

func (m MockEventMsg) Faces() ([]types.Face, int) {
	return []types.Face{}, 0
}

func (m MockEventMsg) Texts() ([]types.Text, int) {
	return []types.Text{}, 0
}

func (m MockEventMsg) ImageFirst() (*types.Image, error) {
	return &types.Image{}, nil
}

func (m MockEventMsg) Images() ([]types.Image, int) {
	return []types.Image{}, 0
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
