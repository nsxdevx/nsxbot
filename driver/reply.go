package driver

import (
	"context"
	"encoding/json"
	"net/http"
)

type HttpReplyer struct {
	Writer http.ResponseWriter
	Cancel context.CancelFunc
}

// http-post https://github.com/botuniverse/onebot-11/blob/master/communication/http-post.md
func (r *HttpReplyer) Reply(data any) error {
	defer r.Cancel()
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = r.Writer.Write(body)
	return err
}

type WSReplyer struct {
	emitter Emitter
	content []byte
}

// ws https://github.com/botuniverse/onebot-11/blob/master/api/hidden.md
func (w *WSReplyer) Reply(data any) error {
	body := struct {
		Context   json.RawMessage `json:"context"`
		Operation any             `json:"operation"`
	}{Context: w.content, Operation: data}
	_, err := w.emitter.Raw(context.Background(), ".handle_quick_operation", body)
	return err
}
