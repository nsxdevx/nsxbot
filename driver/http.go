package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/atopos31/nsxbot/types"
)

type HttpDriver struct {
	*HttpEmitter
	*HttpListener
}

func NewHttpDriver(listener *HttpListener, emitter *HttpEmitter) *HttpDriver {
	return &HttpDriver{
		HttpEmitter:  emitter,
		HttpListener: listener,
	}
}

type HttpListener struct {
	mux  *http.ServeMux
	addr string
	log  *slog.Logger
}

type HttpLIstenerOption func(*HttpListener)

func NewHttpListener(addr string, opts ...HttpLIstenerOption) *HttpListener {
	httpListener := &HttpListener{
		mux:  http.NewServeMux(),
		addr: addr,
		log:  slog.Default().With("HttpListener", addr),
	}
	for _, opt := range opts {
		opt(httpListener)
	}
	return httpListener
}

func (l *HttpListener) Listen(ctx context.Context, eventChan chan<- types.Event) error {
	l.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		content, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		event, err := contentToEvent(content)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			l.log.Error("invalid event", "err", err)
			return
		}
		if event.PostType == types.POST_TYPE_MESSAGE {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			event.Replyer = &types.Replyer{
				Ctx:    ctx,
				Writer: w,
			}
			eventChan <- event
			<-event.Replyer.Ctx.Done()
		} else {
			eventChan <- event
		}

	})
	l.log.Info("Http listener start...", "addr", l.addr)
	server := &http.Server{Addr: l.addr, Handler: l.mux}
	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		server.Shutdown(shutdownCtx)
	}()
	return server.ListenAndServe()
}

type HttpEmitter struct {
	client *http.Client
	url    string
	log    *slog.Logger
}

type HttpEmitterOption func(*HttpEmitter)

func NewHttpEmitter(url string, opts ...HttpEmitterOption) *HttpEmitter {
	httpEmitter := &HttpEmitter{
		client: http.DefaultClient,
		url:    url,
		log:    slog.Default().With("HttpEmitter", url),
	}
	for _, opt := range opts {
		opt(httpEmitter)
	}
	return httpEmitter
}

func (e *HttpEmitter) Raw(ctx context.Context, action types.Action, params any) ([]byte, error) {
	reqbody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.url+"/"+action, bytes.NewBuffer(reqbody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (e *HttpEmitter) GetLoginInfo(ctx context.Context) (*types.LoginInfo, error) {
	return httpAction[any, types.LoginInfo](ctx, e.client, e.url, types.ACTION_GET_LOGIN_INFO, nil)
}

func httpAction[P any, R any](ctx context.Context, client *http.Client, baseurl string, action string, params P) (*R, error) {
	reqbody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseurl+"/"+action, bytes.NewBuffer(reqbody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status error code: %v", res.StatusCode)
	}
	var resp Response[R]
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
