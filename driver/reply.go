package driver

import (
	"context"
	"encoding/json"
	"net/http"
)

type HttpReplyer struct {
	Ctx    context.Context
	Writer http.ResponseWriter
	Cancel context.CancelFunc
}

func (r *HttpReplyer) Reply(data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = r.Writer.Write(body)
	r.Cancel()
	return err
}