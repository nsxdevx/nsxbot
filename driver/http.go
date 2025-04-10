package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/atopos31/nsxbot/types"
)

type DriverHttp struct {
	*EmitterHttp
	*ListenerHttp
}

func NewDriverHttp(listenAddr string, emitterUrl string) *DriverHttp {
	return &DriverHttp{
		EmitterHttp:  NewEmitterHttp(emitterUrl),
		ListenerHttp: NewListenerHttp(listenAddr),
	}
}

type ListenerHttp struct {
	mux  *http.ServeMux
	addr string
	log  *slog.Logger
}

type ListenerHttpOption func(*ListenerHttp)

func NewListenerHttp(addr string, opts ...ListenerHttpOption) *ListenerHttp {
	ListenerHttp := &ListenerHttp{
		mux:  http.NewServeMux(),
		addr: addr,
		log:  slog.Default().WithGroup("[NSXBOT]"),
	}
	for _, opt := range opts {
		opt(ListenerHttp)
	}
	return ListenerHttp
}

func (l *ListenerHttp) Listen(ctx context.Context, eventChan chan<- types.Event) error {
	l.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		content, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			l.log.Error("Read body error", "err", err)
			return
		}
		event, err := contentToEvent(content)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			l.log.Error("Invalid event", "err", err)
			return
		}
		if slices.Contains(event.Types, types.POST_TYPE_MESSAGE) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			event.Replyer = &types.Replyer{
				Ctx:    ctx,
				Writer: w,
				Cancel: cancel,
			}
			eventChan <- event
			<-event.Replyer.Ctx.Done()
		} else {
			eventChan <- event
		}

	})
	l.log.Info("Http listener start... ", "addr", l.addr)
	server := &http.Server{Addr: l.addr, Handler: l.mux}
	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			l.log.Error("Http server shutdown error", "err", err)
			return
		}
	}()
	return server.ListenAndServe()
}

type EmitterHttp struct {
	client *http.Client
	url    string
	log    *slog.Logger
}

type EmitterHttpOption func(*EmitterHttp)

func NewEmitterHttp(url string, opts ...EmitterHttpOption) *EmitterHttp {
	EmitterHttp := &EmitterHttp{
		client: http.DefaultClient,
		url:    url,
		log:    slog.Default().With("EmitterHttp", url),
	}
	for _, opt := range opts {
		opt(EmitterHttp)
	}
	return EmitterHttp
}

func (e *EmitterHttp) Raw(ctx context.Context, action types.Action, params any) ([]byte, error) {
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
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (e *EmitterHttp) SendPvtMsg(ctx context.Context, userId int64, msg types.MeaasgeChain) (*types.SendMsgRes, error) {
	return httpAction[types.SendPrivateMsgReq, types.SendMsgRes](ctx, e.client, e.url, types.ACTION_SEND_PRIVATE_MSG, types.SendPrivateMsgReq{
		UserId:  userId,
		Message: msg,
	})
}
func (e *EmitterHttp) GetLoginInfo(ctx context.Context) (*types.LoginInfo, error) {
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
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status error code: %v", res.StatusCode)
	}
	var resp Response[R]
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}
	if resp.Status == "failed" {
		return nil, fmt.Errorf("action %s failed, retcode: %d, plase see onebot logs", action, resp.RetCode)
	}
	return &resp.Data, nil
}
