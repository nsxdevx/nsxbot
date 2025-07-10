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

	"github.com/nsxdevx/nsxbot/driver"
	"github.com/nsxdevx/nsxbot/event"
	"github.com/nsxdevx/nsxbot/nlog"
)

type HandlerEnd[T any] struct {
	fillers  FilterChain[T]
	handlers HandlersChain[T]
}

type EventHandler[T any] struct {
	selfIds []int64
	Composer[T]
	handlerEnds []HandlerEnd[T]
	log         *slog.Logger
}

func (h *EventHandler[T]) selfs() ([]int64, bool) {
	return h.selfIds, len(h.selfIds) != 0
}

func (h *EventHandler[T]) infos() []string {
	var infos []string
	for _, handlerEnd := range h.handlerEnds {
		var info string
		info += handlerEnd.fillers.debug()
		handler := runtime.FuncForPC(reflect.ValueOf(handlerEnd.handlers[len(handlerEnd.handlers)-1]).Pointer()).Name()
		handler = strings.TrimPrefix(handler, "main.main.")
		info += handler
		infos = append(infos, info)
	}
	return infos
}

func (h *EventHandler[T]) consume(ctx context.Context, emitter driver.Emitter, event event.Event) error {
	var msg T
	if err := json.Unmarshal(event.RawData, &msg); err != nil {
		return err
	}
	h.log.Debug("Consumed", "types", event.Types, "time", event.Time, "selfId", event.SelfId)
	for _, handlerEnd := range h.handlerEnds {
		go func() {
			for _, filter := range handlerEnd.fillers {
				if !filter(msg) {
					return
				}
			}
			h.log.Debug("Handled", "types", event.Types, "time", event.Time, "selfId", event.SelfId, "filter", handlerEnd.fillers.debug())
			nsxctx := NewContext(ctx, emitter, event.Time, event.SelfId, msg, event.Replyer)
			nsxctx.handlers = handlerEnd.handlers
			nsxctx.Next()
		}()
	}
	return nil
}

type consumer interface {
	selfs() ([]int64, bool)
	infos() []string
	consume(ctx context.Context, emitter driver.Emitter, event event.Event) error
}

func SubEvent[T any](engine *Engine, eventype string, selfIds ...int64) *EventHandler[T] {
	handler := &EventHandler[T]{
		log: engine.log,
	}
	// root a pointer to the beginning of the middleware chain
	handler.root = handler
	handler.Use(Recovery[T]())

	engine.consumers[eventype] = handler
	return handler
}

// start handler all self event
func OnEvent[T event.Eventer](engine *Engine) *EventHandler[T] {
	var eventer T
	return SubEvent[T](engine, eventer.Type())
}

// start handler evnet by selfIds
func OnSelfsEvent[T event.Eventer](engine *Engine, selfIds ...int64) *EventHandler[T] {
	var eventer T
	return SubEvent[T](engine, eventer.Type(), selfIds...)
}

type Engine struct {
	listener    driver.Listener
	emitterMux  driver.EmitterMux
	taskLen     int
	consumerNum int
	consumers   map[string]consumer
	log         *slog.Logger
}

func Default(driver driver.Driver) *Engine {
	return &Engine{
		listener:    driver,
		emitterMux:  driver,
		taskLen:     10,
		consumerNum: runtime.NumCPU(),
		consumers:   make(map[string]consumer),
		log:         nlog.Logger(),
	}
}

func New(listener driver.Listener, emitterMux driver.EmitterMux) *Engine {
	return &Engine{
		listener:    listener,
		emitterMux:  emitterMux,
		taskLen:     10,
		consumerNum: runtime.NumCPU(),
		consumers:   make(map[string]consumer),
		log:         nlog.Logger(),
	}
}

func (e *Engine) SetTaskLen(taskLen int) {
	e.taskLen = taskLen
}

func (e *Engine) SetConsumerNum(consumerNum int) {
	e.consumerNum = consumerNum
}

func (e *Engine) debug() {
	e.log.Info("Engine", "taskLen", e.taskLen, "consumerGoruntineNum", e.consumerNum)
	e.log.Info("Consumers", "num", len(e.consumers))
	for t, consumer := range e.consumers {
		for _, info := range consumer.infos() {
			chain := "onebot->"
			if selfIds, ok := consumer.selfs(); ok {
				chain += fmt.Sprintf("selfId:%v->", selfIds)
			} else {
				chain += "all->"
			}
			chain += t + "->" + info
			e.log.Info("Consumer", "chain", chain)
		}
	}
}

func (e *Engine) consumerStart(ctx context.Context, task <-chan event.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-task:
			e.log.Debug("Received", "types", event.Types, "time", event.Time, "selfId", event.SelfId)
			for _, Type := range event.Types {
				if consumer, ok := e.consumers[Type]; ok {
					if selfIds, ok := consumer.selfs(); ok && !slices.Contains(selfIds, event.SelfId) {
						continue
					}
					emitter, err := e.emitterMux.GetEmitter(event.SelfId)
					if err != nil {
						e.log.Error("GetEmitter error", "error", err)
						continue
					}
					if err := consumer.consume(context.Background(), emitter, event); err != nil {
						e.log.Error("Consume error", "error", err)
						continue
					}
				}
			}
		}
	}
}

func (e *Engine) Run(ctx context.Context) {
	e.debug()
	task := make(chan event.Event, e.taskLen)
	for range e.consumerNum {
		go e.consumerStart(ctx, task)
	}
	if nlog.Leveler == slog.LevelDebug {
		e.log.Warn("Run in debug mode, please set env NSX_MODE=release or nlog.SetLevel(slog.LevelInfo) to disable debug mode.")
	}
	if err := e.listener.Listen(ctx, task); err != nil {
		panic(err)
	}
}
