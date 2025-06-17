package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nsxdevx/nsxbot/event"
	"github.com/nsxdevx/nsxbot/nlog"
	"github.com/nsxdevx/nsxbot/schema"
	"github.com/nsxdevx/nsxbot/types"
	"github.com/tidwall/gjson"
)

type echoStore struct {
	mu    sync.RWMutex
	echos map[int64]chan Response[json.RawMessage]
}

func (e *echoStore) Receive(selfId int64, data []byte) error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	echoChan, ok := e.echos[selfId]
	if !ok {
		return fmt.Errorf("no echo store for selfId %d", selfId)
	}
	var echo Response[json.RawMessage]
	if err := json.Unmarshal(data, &echo); err != nil {
		return err
	}
	echoChan <- echo
	return nil
}

func (e *echoStore) GetEcho(selfId int64) chan Response[json.RawMessage] {
	e.mu.Lock()
	defer e.mu.Unlock()
	echoChan, ok := e.echos[selfId]
	if ok {
		return echoChan
	}
	echoChan = make(chan Response[json.RawMessage], 10)
	e.echos[selfId] = echoChan
	return echoChan
}

type WSnode struct {
	Url   string
	Token string
}

type WSClient struct {
	*WSEmittersMux
	echoStore  *echoStore
	nodes      []WSnode
	log        *slog.Logger
	retryDelay time.Duration
}

func NewWSClient(retryDelay time.Duration, nodes ...WSnode) *WSClient {
	return &WSClient{
		WSEmittersMux: &WSEmittersMux{
			emitters:         make(map[int64]Emitter),
			connectCallbacks: make(map[int64]func(Emitter)),
			log:              nlog.Logger(),
		},
		echoStore: &echoStore{
			echos: make(map[int64]chan Response[json.RawMessage]),
		},
		nodes:      nodes,
		log:        nlog.Logger(),
		retryDelay: retryDelay,
	}
}

func (ws *WSClient) Listen(ctx context.Context, eventChan chan<- event.Event) error {
	for _, node := range ws.nodes {
		go func(ctx context.Context) {
			ticker := time.NewTicker(ws.retryDelay)
			url := "ws://" + node.Url
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					header := make(http.Header, 1)
					header.Set("Authorization", "Bearer "+node.Token)
					c, _, err := websocket.DefaultDialer.Dial(url, header)
					if err != nil {
						ws.log.Error("Dial", "err", err)
						continue
					}
					defer func() {
						if err := c.Close(); err != nil {
							ws.log.Error("Close", "err", err)
						}
					}()
					// connSelfId only be use in meta_event and echoStore
					var connSelfId int64
					for {
						_, content, err := c.ReadMessage()
						if err != nil {
							ws.log.Error("Read", "err", err)
							ws.RemoveEmitter(connSelfId)
							break
						}
						go func() {
							if gjson.Get(string(content), "echo").Exists() {
								if err := ws.echoStore.Receive(connSelfId, content); err != nil {
									ws.log.Error("Receive echo", "err", err)
								}
								return
							}

							botevent, err := contentToEvent(content)
							if err != nil {
								ws.log.Error("Invalid event", "err", err)
								return
							}

							emitter := NewEmitterWS(botevent.SelfId, c, ws.echoStore.GetEcho(botevent.SelfId))

							if slices.Contains(botevent.Types, event.EVENT_META) {
								connSelfId = botevent.SelfId
								ws.AddEmitter(connSelfId, emitter)
							}

							if slices.Contains(botevent.Types, event.EVENT_MESSAGE) || slices.Contains(botevent.Types, event.EVENT_REQUEST) {
								botevent.Replyer = &WSReplyer{
									content: content,
									emitter: emitter,
								}
							}
							eventChan <- botevent
						}()
					}
				}
			}
		}(ctx)
	}
	<-ctx.Done()
	return nil
}

type WServer struct {
	*WSEmittersMux
	echoStore *echoStore
	url       url.URL
	token     string
	log       *slog.Logger
}

type WServerOption func(*WServer)

func WSerevrWithToken(token string) WServerOption {
	return func(ws *WServer) {
		ws.token = token
	}
}

func NewWSverver(host string, path string, opts ...WServerOption) *WServer {
	ws := &WServer{
		WSEmittersMux: &WSEmittersMux{
			emitters:         make(map[int64]Emitter),
			connectCallbacks: make(map[int64]func(Emitter)),
			log:              nlog.Logger(),
		},
		echoStore: &echoStore{
			echos: make(map[int64]chan Response[json.RawMessage]),
		},
		url: url.URL{
			Scheme: "ws",
			Host:   host,
			Path:   path,
		},
		log: nlog.Logger(),
	}
	for _, opt := range opts {
		opt(ws)
	}
	return ws
}

func (ws *WServer) Listen(ctx context.Context, eventChan chan<- event.Event) error {
	mux := http.NewServeMux()
	mux.HandleFunc(ws.url.Path, func(w http.ResponseWriter, r *http.Request) {
		if err := ws.auth(r); err != nil {
			ws.log.Error("Invalid token", "err", err)
			return
		}
		var upgrader = websocket.Upgrader{}
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			ws.log.Error("Upgrade", "err", err)
			return
		}
		defer func() {
			if err := c.Close(); err != nil {
				ws.log.Error("Close", "err", err)
			}
		}()
		// connSelfId only be use in meta_event and echoStore
		var connSelfId int64
		for {
			_, content, err := c.ReadMessage()
			if err != nil {
				ws.log.Error("Read", "err", err)
				ws.RemoveEmitter(connSelfId)
				break
			}

			go func() {
				if gjson.Get(string(content), "echo").Exists() {
					if err := ws.echoStore.Receive(connSelfId, content); err != nil {
						ws.log.Error("Receive echo", "err", err)
					}
					return
				}

				botevent, err := contentToEvent(content)
				if err != nil {
					ws.log.Error("Invalid event", "err", err)
					return
				}

				emitter := NewEmitterWS(botevent.SelfId, c, ws.echoStore.GetEcho(botevent.SelfId))

				if slices.Contains(botevent.Types, event.EVENT_META) {
					connSelfId = botevent.SelfId
					ws.AddEmitter(connSelfId, emitter)
				}

				if slices.Contains(botevent.Types, event.EVENT_MESSAGE) || slices.Contains(botevent.Types, event.EVENT_REQUEST) {
					botevent.Replyer = &WSReplyer{
						content: content,
						emitter: emitter,
					}
				}
				eventChan <- botevent
			}()
		}
	})
	ws.log.Info("WS listener start... ", "addr", ws.url.Host)
	server := &http.Server{Addr: ws.url.Host, Handler: mux}
	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			ws.log.Error("WS server shutdown error", "err", err)
			return
		}
	}()
	return server.ListenAndServe()
}

func (ws *WServer) auth(r *http.Request) error {
	if len(ws.token) == 0 {
		return nil
	}
	token := r.Header.Get("Authorization")
	if strings.EqualFold("Bearer "+ws.token, token) {
		return nil
	}
	return fmt.Errorf("invalid token")
}

type WSEmittersMux struct {
	mu               sync.RWMutex
	callmu           sync.RWMutex
	emitters         map[int64]Emitter
	connectCallbacks map[int64]func(Emitter)
	onClose          func(int64)
	log              *slog.Logger
}

func (ws *WSEmittersMux) AddEmitter(selfId int64, emitter Emitter) {
	info, err := emitter.GetVersionInfo(context.Background())
	if err != nil {
		ws.log.Warn("GetVersionInfo error", "error", err, "selfId", selfId)
		info = &types.VersionInfo{
			AppName:         "unknown",
			ProtocolVersion: "unknown",
			AppVersion:      "unknown",
		}
	}
	ws.mu.RLock()
	if _, ok := ws.emitters[selfId]; ok {
		ws.mu.RUnlock()
		return
	}
	ws.mu.RUnlock()

	ws.log.Info("NewEmitterWS", "selfId", selfId, "AppName", info.AppName, "ProtocolVersion", info.ProtocolVersion, "AppVersion", info.AppVersion)

	ws.mu.Lock()
	ws.emitters[selfId] = emitter
	ws.mu.Unlock()
	if callBack, ok := ws.connectCallbacks[selfId]; ok {
		go callBack(emitter)
	}
}

func (ws *WSEmittersMux) OnConnect(selfId int64, callback func(Emitter)) {
	ws.callmu.Lock()
	defer ws.callmu.Unlock()
	ws.connectCallbacks[selfId] = callback
}

func (ws *WSEmittersMux) RemoveEmitter(selfId int64) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	delete(ws.emitters, selfId)
	if ws.onClose != nil {
		go ws.onClose(selfId)
	}
}

func (ws *WSEmittersMux) OnClose(callback func(int64)) {
	ws.onClose = callback
}

func (ws *WSEmittersMux) GetEmitter(selfId int64) (Emitter, error) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	emitter, ok := ws.emitters[selfId]
	if !ok {
		return nil, fmt.Errorf("emitter not found")
	}
	return emitter, nil
}

type EmitterWS struct {
	mu     sync.Mutex
	conn   *websocket.Conn
	echo   chan Response[json.RawMessage]
	selfId int64
	log    *slog.Logger
}

func NewEmitterWS(selfId int64, conn *websocket.Conn, echo chan Response[json.RawMessage]) *EmitterWS {
	return &EmitterWS{
		conn:   conn,
		echo:   echo,
		selfId: selfId,
		log:    nlog.Logger(),
	}
}

func (e *EmitterWS) SendPvtMsg(ctx context.Context, userId int64, msg schema.MessageChain) (*types.SendMsgRes, error) {
	e.mu.Lock()
	echoId, err := wsAction(e.conn, ACTION_SEND_PRIVATE_MSG, types.SendPrivateMsgReq{
		UserId:  userId,
		Message: msg,
	})
	if err != nil {
		e.mu.Unlock()
		return nil, err
	}
	e.mu.Unlock()
	return wsWait[types.SendMsgRes](ctx, echoId, e.echo)
}

func (e *EmitterWS) SendGrMsg(ctx context.Context, groupId int64, msg schema.MessageChain) (*types.SendMsgRes, error) {
	e.mu.Lock()
	echoId, err := wsAction(e.conn, ACTION_SEND_GROUP_MSG, types.SendGrMsgReq{
		GroupId: groupId,
		Message: msg,
	})
	if err != nil {
		e.mu.Unlock()
		return nil, err
	}
	e.mu.Unlock()
	return wsWait[types.SendMsgRes](ctx, echoId, e.echo)
}

func (e *EmitterWS) DelMsg(ctx context.Context, msgId int) error {
	e.mu.Lock()
	echoId, err := wsAction[any](e.conn, ACTION_DELETE_MSG, types.DelMsgReq{
		MessageId: msgId,
	})
	if err != nil {
		e.mu.Unlock()
		return err
	}
	e.mu.Unlock()
	_, err = wsWait[any](ctx, echoId, e.echo)
	return err
}

func (e *EmitterWS) GetMsg(ctx context.Context, msgId int) (*types.GetMsgRes, error) {
	e.mu.Lock()
	echoId, err := wsAction(e.conn, ACTION_GET_MSG, types.GetMsgReq{
		MessageId: msgId,
	})
	if err != nil {
		e.mu.Unlock()
		return nil, err
	}
	e.mu.Unlock()
	return wsWait[types.GetMsgRes](ctx, echoId, e.echo)
}

func (e *EmitterWS) GetLoginInfo(ctx context.Context) (*types.LoginInfo, error) {
	e.mu.Lock()
	echoId, err := wsAction[any](e.conn, ACTION_GET_LOGIN_INFO, nil)
	if err != nil {
		e.mu.Unlock()
		return nil, err
	}
	e.mu.Unlock()
	return wsWait[types.LoginInfo](ctx, echoId, e.echo)
}

func (e *EmitterWS) GetStrangerInfo(ctx context.Context, userId int64, noCache bool) (*types.StrangerInfo, error) {
	e.mu.Lock()
	echoId, err := wsAction(e.conn, ACTION_GET_STRANGER_INFO, types.GetStrangerInfo{
		UserId:  userId,
		NoCache: noCache,
	})
	if err != nil {
		e.mu.Unlock()
		return nil, err
	}
	e.mu.Unlock()
	return wsWait[types.StrangerInfo](ctx, echoId, e.echo)
}

func (e *EmitterWS) GetStatus(ctx context.Context) (*types.Status, error) {
	e.mu.Lock()
	echoId, err := wsAction[any](e.conn, ACTION_GET_STATUS, nil)
	if err != nil {
		e.mu.Unlock()
		return nil, err
	}
	e.mu.Unlock()
	return wsWait[types.Status](ctx, echoId, e.echo)
}

func (e *EmitterWS) GetVersionInfo(ctx context.Context) (*types.VersionInfo, error) {
	e.mu.Lock()
	echoId, err := wsAction[any](e.conn, ACTION_GET_VERSION_INFO, nil)
	if err != nil {
		e.mu.Unlock()
		return nil, err
	}
	e.mu.Unlock()
	return wsWait[types.VersionInfo](ctx, echoId, e.echo)
}

func (e *EmitterWS) GetSelfId(ctx context.Context) (int64, error) {
	return e.selfId, nil
}

func (e *EmitterWS) SetFriendAddRequest(ctx context.Context, flag string, approve bool, remark string) error {
	e.mu.Lock()
	echoId, err := wsAction(e.conn, ACTION_SET_FRIEND_ADD_REQUEST, types.FriendAddReq{
		Flag:    flag,
		Approve: approve,
		Remark:  remark,
	})
	if err != nil {
		e.mu.Unlock()
		return err
	}
	e.mu.Unlock()
	_, err = wsWait[any](ctx, echoId, e.echo)
	return err
}

func (e *EmitterWS) SetGroupAddRequest(ctx context.Context, flag string, approve bool, reason string) error {
	e.mu.Lock()
	echoId, err := wsAction(e.conn, ACTION_SET_GROUP_ADD_REQUEST, types.GroupAddReq{
		Flag:    flag,
		Approve: approve,
		Reason:  reason,
	})
	if err != nil {
		e.mu.Unlock()
		return err
	}
	e.mu.Unlock()
	_, err = wsWait[any](ctx, echoId, e.echo)
	return err
}

func (e *EmitterWS) SetGroupSpecialTitle(ctx context.Context, groupId int64, userId int64, specialTitle string, duration int) error {
	e.mu.Lock()
	echoId, err := wsAction(e.conn, ACTION_SET_GROUP_SPECIAL_TITLE, types.SpecialTitleReq{
		GroupId:      groupId,
		UserId:       userId,
		SpecialTitle: specialTitle,
	})
	if err != nil {
		e.mu.Unlock()
		return err
	}
	e.mu.Unlock()
	_, err = wsWait[any](ctx, echoId, e.echo)
	return err
}

func (e *EmitterWS) Raw(ctx context.Context, action Action, params any) ([]byte, error) {
	e.mu.Lock()
	echoId, err := wsAction(e.conn, action, params)
	if err != nil {
		e.mu.Unlock()
		return nil, err
	}
	e.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case echo := <-e.echo:
			if !strings.EqualFold(echoId, echo.Echo) {
				e.echo <- echo
				continue
			}
			return json.Marshal(echo)
		}
	}
}

func wsAction[P any](conn *websocket.Conn, action string, params P) (string, error) {
	echoid := uuid.New().String()
	return echoid, conn.WriteJSON(Request[P]{
		Action: action,
		Echo:   echoid,
		Params: params,
	})
}

func wsWait[R any](ctx context.Context, echoId string, echoChan chan Response[json.RawMessage]) (*R, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case echo := <-echoChan:
			if !strings.EqualFold(echoId, echo.Echo) {
				echoChan <- echo
				continue
			}
			if strings.EqualFold("failed", echo.Status) {
				return nil, fmt.Errorf("action failed, rawdata: %x, please see onebot logs", echo.Status)
			}
			var res R
			if err := json.Unmarshal(echo.Data, &res); err != nil {
				return nil, err
			}
			return &res, nil
		}
	}
}
