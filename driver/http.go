package driver

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/nsxdevx/nsxbot/event"
	"github.com/nsxdevx/nsxbot/nlog"
	"github.com/nsxdevx/nsxbot/schema"
	"github.com/nsxdevx/nsxbot/types"
)

type DriverHttp struct {
	*EmitterMuxHttp
	*ListenerHttp
}

func NewDriverHttp(listenAddr string, emitterUrl ...string) *DriverHttp {
	return &DriverHttp{
		EmitterMuxHttp: NewEmitterMuxHttp(emitterUrl...),
		ListenerHttp:   NewListenerHttp(listenAddr),
	}
}

type ListenerHttp struct {
	mux          *http.ServeMux
	addr         string
	token        string
	replyTimeout time.Duration
	log          *slog.Logger
}

type ListenerHttpOption func(*ListenerHttp)

func ListenerHttpWithTimeout(timeout time.Duration) ListenerHttpOption {
	return func(l *ListenerHttp) {
		l.replyTimeout = timeout
	}
}

func ListenerHttpWithToken(token string) ListenerHttpOption {
	return func(l *ListenerHttp) {
		l.token = token
	}
}

func NewListenerHttp(addr string, opts ...ListenerHttpOption) *ListenerHttp {
	ListenerHttp := &ListenerHttp{
		mux:          http.NewServeMux(),
		addr:         addr,
		replyTimeout: 1 * time.Second,
		log:          nlog.Logger(),
	}
	for _, opt := range opts {
		opt(ListenerHttp)
	}
	return ListenerHttp
}

func (l *ListenerHttp) Listen(ctx context.Context, eventChan chan<- event.Event) error {
	l.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content, err := l.auth(w, r)
		if err != nil {
			l.log.Error("Invalid content", "err", err)
			return
		}
		botevent, err := contentToEvent(content)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			l.log.Error("Invalid event", "err", err)
			return
		}
		if slices.Contains(botevent.Types, event.EVENT_MESSAGE) || slices.Contains(botevent.Types, event.EVENT_NOTICE) {
			ctx, cancel := context.WithTimeout(context.Background(), l.replyTimeout)
			botevent.Replyer = &HttpReplyer{
				Writer: w,
				Cancel: cancel,
			}
			eventChan <- botevent
			<-ctx.Done()
		} else {
			eventChan <- botevent
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

func (l *ListenerHttp) auth(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil, fmt.Errorf("method not allowed")
	}
	content, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}
	if len(l.token) != 0 {
		sign := r.Header.Get("X-Signature")
		if len(sign) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return nil, fmt.Errorf("invalid token")
		}
		mac := hmac.New(sha1.New, []byte(l.token))
		mac.Write(content)
		if sign != "sha1="+hex.EncodeToString(mac.Sum(nil)) {
			w.WriteHeader(http.StatusForbidden)
			return nil, fmt.Errorf("invalid token")
		}
	}
	return content, nil
}

type EmitterMuxHttp struct {
	mu       sync.RWMutex
	emitters map[int64]Emitter
	log      *slog.Logger
}

func NewEmitterMuxHttpSets(emitterhttps ...*EmitterHttp) *EmitterMuxHttp {
	emitters := make(map[int64]Emitter, len(emitterhttps))
	for _, emitter := range emitterhttps {
		id, err := emitter.GetSelfId(context.Background())
		if err != nil {
			panic(err)
		}
		emitters[id] = emitter
	}
	return &EmitterMuxHttp{
		emitters: emitters,
		log:      nlog.Logger(),
	}
}

func NewEmitterMuxHttp(urls ...string) *EmitterMuxHttp {
	mux := &EmitterMuxHttp{
		emitters: make(map[int64]Emitter),
		log:      nlog.Logger(),
	}
	for _, url := range urls {
		go func() {
			emitter := NewEmitterHttp(url)
			selfId, err := emitter.GetSelfId(context.Background())
			if err != nil {
				panic(err)
			}
			mux.AddEmitter(selfId, emitter)
		}()
	}
	return mux
}

func (m *EmitterMuxHttp) AddEmitter(selfId int64, emitter Emitter) {
	info, err := emitter.GetVersionInfo(context.Background())
	if err != nil {
		m.log.Warn("GetVersionInfo error", "error", err, "selfId", selfId)
	} else {
		m.log.Info("NewEmitterHttp", "selfId", selfId, "appName", info.AppName, "protocolVersion", info.ProtocolVersion, "appVersion", info.AppVersion)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.emitters[selfId] = emitter
}

func (m *EmitterMuxHttp) RemoveEmitter(selfId int64) {
	delete(m.emitters, selfId)
}

func (m *EmitterMuxHttp) GetEmitter(selfId int64) (Emitter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	emitter, ok := m.emitters[selfId]
	if !ok {
		return nil, fmt.Errorf("emitter not found")
	}
	return emitter, nil
}

type EmitterHttp struct {
	client *http.Client
	url    string
	token  string
	selfId *int64
	log    *slog.Logger
}

type EmitterHttpOption func(*EmitterHttp)

func NewEmitterHttp(url string, opts ...EmitterHttpOption) *EmitterHttp {
	EmitterHttp := &EmitterHttp{
		client: http.DefaultClient,
		url:    url,
		log:    nlog.Logger(),
	}
	for _, opt := range opts {
		opt(EmitterHttp)
	}
	return EmitterHttp
}

// Set selfId to EmitterHttp, instand of get from GetLoginInfo
func WithEmitterHttpSelfId(selfId int64) EmitterHttpOption {
	return func(e *EmitterHttp) {
		e.selfId = &selfId
	}
}

func WithEmitterHttpToken(token string) EmitterHttpOption {
	return func(e *EmitterHttp) {
		e.token = token
	}
}

func (e *EmitterHttp) Raw(ctx context.Context, action Action, params any) ([]byte, error) {
	reqbody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.url+"/"+action, bytes.NewBuffer(reqbody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.token)
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

func (e *EmitterHttp) SendPvtMsg(ctx context.Context, userId int64, msg schema.MessageChain) (*types.SendMsgRes, error) {
	return httpAction[types.SendPrivateMsgReq, types.SendMsgRes](ctx, e.client, e.token, e.url, ACTION_SEND_PRIVATE_MSG, types.SendPrivateMsgReq{
		UserId:  userId,
		Message: msg,
	})
}

func (e *EmitterHttp) SendGrMsg(ctx context.Context, groupId int64, msg schema.MessageChain) (*types.SendMsgRes, error) {
	return httpAction[types.SendGrMsgReq, types.SendMsgRes](ctx, e.client, e.token, e.url, ACTION_SEND_GROUP_MSG, types.SendGrMsgReq{
		GroupId: groupId,
		Message: msg,
	})
}

func (e *EmitterHttp) GetMsg(ctx context.Context, msgId int) (*types.GetMsgRes, error) {
	return httpAction[types.GetMsgReq, types.GetMsgRes](ctx, e.client, e.token, e.url, ACTION_GET_MSG, types.GetMsgReq{
		MessageId: msgId,
	})
}

func (e *EmitterHttp) DelMsg(ctx context.Context, messageId int) error {
	_, err := httpAction[types.DelMsgReq, any](ctx, e.client, e.token, e.url, ACTION_DELETE_MSG, types.DelMsgReq{
		MessageId: messageId,
	})
	return err
}

func (e *EmitterHttp) GetLoginInfo(ctx context.Context) (*types.LoginInfo, error) {
	return httpAction[any, types.LoginInfo](ctx, e.client, e.token, e.url, ACTION_GET_LOGIN_INFO, nil)
}

func (e *EmitterHttp) GetStrangerInfo(ctx context.Context, userId int64, noCache bool) (*types.StrangerInfo, error) {
	return httpAction[types.GetStrangerInfo, types.StrangerInfo](ctx, e.client, e.token, e.url, ACTION_GET_STRANGER_INFO, types.GetStrangerInfo{
		UserId:  userId,
		NoCache: noCache,
	})
}

func (e *EmitterHttp) GetStatus(ctx context.Context) (*types.Status, error) {
	return httpAction[any, types.Status](ctx, e.client, e.token, e.url, ACTION_GET_STATUS, nil)
}

func (e *EmitterHttp) GetVersionInfo(ctx context.Context) (*types.VersionInfo, error) {
	return httpAction[any, types.VersionInfo](ctx, e.client, e.token, e.url, ACTION_GET_VERSION_INFO, nil)
}

func (e *EmitterHttp) GetSelfId(ctx context.Context) (int64, error) {
	if e.selfId != nil {
		return *e.selfId, nil
	}
	e.log.Warn("SelfId is nil, try get from GetLoginInfo", "url", e.url)
	info, err := e.GetLoginInfo(ctx)
	if err != nil {
		return 0, err
	}
	e.selfId = &info.UserId
	return *e.selfId, nil
}

func (e *EmitterHttp) SetFriendAddRequest(ctx context.Context, flag string, approve bool, remark string) error {
	_, err := httpAction[types.FriendAddReq, any](ctx, e.client, e.token, e.url, ACTION_SET_FRIEND_ADD_REQUEST, types.FriendAddReq{
		Flag:    flag,
		Approve: approve,
		Remark:  remark,
	})
	return err
}

func (e *EmitterHttp) SetGroupAddRequest(ctx context.Context, flag string, approve bool, reason string) error {
	_, err := httpAction[types.GroupAddReq, any](ctx, e.client, e.token, e.url, ACTION_SET_GROUP_ADD_REQUEST, types.GroupAddReq{
		Flag:    flag,
		Approve: approve,
		Reason:  reason,
	})
	return err
}

func (e *EmitterHttp) SetGroupSpecialTitle(ctx context.Context, groupId int64, userId int64, specialTitle string, duration int) error {
	_, err := httpAction[types.SpecialTitleReq, any](ctx, e.client, e.token, e.url, ACTION_SET_GROUP_SPECIAL_TITLE, types.SpecialTitleReq{
		GroupId:      groupId,
		UserId:       userId,
		SpecialTitle: specialTitle,
	})

	return err
}

func httpAction[P any, R any](ctx context.Context, client *http.Client, token string, baseurl string, action string, params P) (*R, error) {
	reqbody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseurl+"/"+action, bytes.NewBuffer(reqbody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status error code: %v", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var resp Response[R]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if strings.EqualFold("failed", resp.Status) {
		return nil, fmt.Errorf("action %s failed, rawdata: %s, please see onebot logs", action, string(body))
	}
	return &resp.Data, nil
}
