package types

import (
	"errors"
)

var (
	ErrNoAvailable = errors.New("no replayer available")
)

type Replayer interface {
	Reply(data any) error
}
