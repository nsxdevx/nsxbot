package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/atopos31/nsxbot/nlog"
	"github.com/atopos31/nsxbot/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

type WServer struct {
	mu       sync.RWMutex
	emitters map[int64]*EmitterWS
	echo     chan Response[json.RawMessage]
	url      url.URL
	log      *slog.Logger
}

func NewWSverver(host string, path string) *WServer {
	return &WServer{
		emitters: make(map[int64]*EmitterWS),
		echo:     make(chan Response[json.RawMessage], 100),
		url: url.URL{
			Scheme: "ws",
			Host:   host,
			Path:   path,
		},
		log: nlog.Logger(),
	}
}

func (ws *WServer) Listen(ctx context.Context, eventChan chan<- types.Event) error {
	mux := http.NewServeMux()
	mux.HandleFunc(ws.url.Path, func(w http.ResponseWriter, r *http.Request) {
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
		var selfId int64
		for {
			_, content, err := c.ReadMessage()
			if err != nil {
				ws.log.Error("Read", "err", err)
				break
			}
			if gjson.Get(string(content), "echo").Exists() {
				var echo Response[json.RawMessage]
				if err := json.Unmarshal(content, &echo); err != nil {
					ws.log.Error("Invalid echo", "err", err)
					continue
				}
				ws.echo <- echo
				continue
			}
			event, err := contentToEvent(content)
			if err != nil {
				ws.log.Error("Invalid event", "err", err)
				continue
			}
			selfId = event.SelfID
			ws.AddEmitter(event.SelfID, NewEmitterWS(event.SelfID, c, ws.echo))
			eventChan <- event
		}
		ws.mu.Lock()
		delete(ws.emitters, selfId)
		defer ws.mu.Lock()
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

func (ws *WServer) AddEmitter(selfId int64, emitter *EmitterWS) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.emitters[selfId] = emitter
}

func (ws *WServer) GetEmitter(selfId int64) (*EmitterWS, error) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	emitter, ok := ws.emitters[selfId]
	if !ok {
		return nil, fmt.Errorf("emitter not found")
	}
	return emitter, nil
}

type EmitterWS struct {
	mu     sync.RWMutex
	conn   *websocket.Conn
	echo   chan Response[json.RawMessage]
	selfId *int64
	log    *slog.Logger
}

func NewEmitterWS(selfId int64, conn *websocket.Conn, echo chan Response[json.RawMessage]) *EmitterWS {
	return &EmitterWS{
		conn:   conn,
		echo:   echo,
		selfId: &selfId,
		log:    nlog.Logger(),
	}
}

func (e *EmitterWS) GetStatus(ctx context.Context) (*types.Status, error) {
	e.mu.Lock()
	echoId, err := wsAction[any](e.conn, Action_GET_STATUS, nil)
	if err != nil {
		e.mu.Unlock()
		return nil, err
	}
	e.mu.Unlock()
	return wsWait[types.Status](ctx, echoId, e.echo)
}

func wsAction[P any](conn *websocket.Conn, action string, params P) (string, error) {
	echoid := uuid.New().String()
	return echoid, conn.WriteJSON(Request[P]{
		Action: action,
		Echo:   echoid,
		Params: params,
	})
}

func wsWait[R any](ctx context.Context, echoId string, echocChan chan Response[json.RawMessage]) (*R, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case echo := <-echocChan:
			if !strings.EqualFold(echoId, echo.Echo) {
				echocChan <- echo
				continue
			}
			var res R
			if err := json.Unmarshal(echo.Data, &res); err != nil {
				return nil, err
			}
			return &res, nil
		}
	}
}
