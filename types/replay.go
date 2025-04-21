package types

import (
	"errors"
)

var (
	ErrNoAvailable = errors.New("no Replyer available")
)

type Replyer interface {
	Reply(data any) error
}
