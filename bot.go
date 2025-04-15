package nsxbot

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"slices"
	"strings"

	"github.com/atopos31/nsxbot/driver"
	"github.com/atopos31/nsxbot/nlog"
	"github.com/atopos31/nsxbot/types"
)

type HandlerEnd[T any] struct {
	fillers  FilterChain[T]
	handlers HandlersChain[T]
}

type EventHandler[T types.Eventer] struct {
	selfIds []int64
	Composer[T]
	handlerEnds []HandlerEnd[T]
}

func (h *EventHandler[T]) selfs() ([]int64, bool) {
	return h.selfIds, len(h.selfIds) != 0
}

func (h *EventHandler[T]) infos() []string {
	var infos []string
	for _, handlerEnd := range h.handlerEnds {
		var info string
		for _, filter := range handlerEnd.fillers {
			info += strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(filter).Pointer()).Name()+"->", "main.main.")
		}
		handler := runtime.FuncForPC(reflect.ValueOf(handlerEnd.handlers[len(handlerEnd.handlers)-1]).Pointer()).Name()
		handler = strings.TrimPrefix(handler, "main.main.")
		info += handler
		infos = append(infos, info)
	}
	return infos
}

func (h *EventHandler[T]) consume(ctx context.Context, emitter driver.Emitter, event types.Event) error {
	var msg T
	if err := json.Unmarshal(event.RawData, &msg); err != nil {
		return err
	}
	for _, handlerEnd := range h.handlerEnds {
		go func() {
			for _, filter := range handlerEnd.fillers {
				if !filter(msg) {
					return
				}
			}
			nsxctx := NewContext(ctx, emitter, event.Time, event.SelfID, msg, event.Replyer)
			nsxctx.handlers = handlerEnd.handlers
			nsxctx.Next()
		}()
	}
	return nil
}

type consumer interface {
	selfs() ([]int64, bool)
	infos() []string
	consume(ctx context.Context, emitter driver.Emitter, event types.Event) error
}

// start handler all self event
func OnEvent[T types.Eventer](engine *Engine) *EventHandler[T] {
	eventHandler := new(EventHandler[T])
	eventHandler.root = eventHandler

	var eventer T
	engine.consumers[eventer.Type()] = eventHandler
	return eventHandler
}

// start handler evnet by selfIds
func OnSelfsEvent[T types.Eventer](engine *Engine, selfIds ...int64) *EventHandler[T] {
	eventHandler := OnEvent[T](engine)
	eventHandler.selfIds = selfIds
	return eventHandler
}

type Engine struct {
	listener    driver.Listener
	emitterMux  driver.EmitterMux
	taskLen     int
	consumerNum int
	consumers   map[types.EventType]consumer
	logger      *slog.Logger
}

func Default(ctx context.Context, driver driver.Driver) *Engine {
	return &Engine{
		listener:    driver,
		emitterMux:  driver,
		taskLen:     10,
		consumerNum: runtime.NumCPU(),
		consumers:   make(map[types.EventType]consumer),
		logger:      nlog.Logger(),
	}
}

func New(ctx context.Context, listener driver.Listener, emitterMux driver.EmitterMux) *Engine {
	return &Engine{
		listener:    listener,
		emitterMux:  emitterMux,
		taskLen:     10,
		consumerNum: runtime.NumCPU(),
		consumers:   make(map[types.EventType]consumer),
		logger:      nlog.Logger(),
	}
}

func (e *Engine) SetTaskLen(taskLen int) {
	e.taskLen = taskLen
}

func (e *Engine) SetConsumerNum(consumerNum int) {
	e.consumerNum = consumerNum
}

func (e *Engine) debug() {
	e.logger.Info("Engine", "taskLen", e.taskLen, "consumerNum", e.consumerNum)
	e.logger.Info("Consumers", "num", len(e.consumers))
	for t, consumer := range e.consumers {
		for _, info := range consumer.infos() {
			chain := "onebot->"
			if selfIds, ok := consumer.selfs(); ok {
				chain += fmt.Sprintf("selfId:%v->", selfIds)
			} else {
				chain += "all->"
			}
			chain += t + "->" + info
			e.logger.Info("Consumer", "chain", chain)
		}
	}
}

func (e *Engine) consumerStart(ctx context.Context, task <-chan types.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-task:
			e.logger.Debug("Received", "event", event.Types, "time", event.Time, "selfID", event.SelfID)
			for _, Type := range event.Types {
				if consumer, ok := e.consumers[Type]; ok {
					if selfIds, ok := consumer.selfs(); ok && !slices.Contains(selfIds, event.SelfID) {
						continue
					}
					emitter, err := e.emitterMux.GetEmitter(event.SelfID)
					if err != nil {
						e.logger.Error("GetEmitter error", "error", err)
						continue
					}
					if err := consumer.consume(ctx, emitter, event); err != nil {
						e.logger.Error("Consume error", "error", err)
						continue
					}
					e.logger.Info("Consumed", "types", event.Types, "time", event.Time, "selfID", event.SelfID)
				}
			}
		}
	}
}

func (e *Engine) Run(ctx context.Context) {
	e.debug()
	task := make(chan types.Event, e.taskLen)
	for range e.consumerNum {
		go e.consumerStart(ctx, task)
	}
	if err := e.listener.Listen(ctx, task); err != nil {
		panic(err)
	}
}
