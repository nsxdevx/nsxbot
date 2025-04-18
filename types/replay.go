package types

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrNoAvailable = errors.New("no replayer available")
)

type Replayer interface {
	Reply(data []byte) error
}

type HttpReplyer struct {
	Ctx    context.Context
	Writer http.ResponseWriter
	Cancel context.CancelFunc
}

func (r *HttpReplyer) Reply(data []byte) error {
	_, err := r.Writer.Write(data)
	r.Cancel()
	return err
}
