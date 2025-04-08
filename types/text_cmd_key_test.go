package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmdKey(t *testing.T) {
	tests := []struct {
		text  string
		key   string
		want  string
		want1 error
	}{
		{"/test", "key", "", errors.New("not enough parts")},
		{"/test key", "key", "", errors.New("not enough parts")},
		{"/test key value", "key", "value", nil},
		{"/test key value targetKey targetValue", "targetKey", "targetValue", nil},
		{"/test key value anotherKey anotherValue", "targetKey", "", errors.New("key not found")},
		{"/test key value TargetKey targetValue", "targetKey", "targetValue", nil},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			text := &Text{Text: tt.text}
			got, got1 := text.CmdKey(tt.key)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}
