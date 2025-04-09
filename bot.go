package nsxbot

import (
	"context"
	"encoding/json"
	"log/slog"
	"runtime"

	"github.com/atopos31/nsxbot/driver"
	"github.com/atopos31/nsxbot/types"
)

type Eventer interface {
	Type() string
}

type HandlerEnd[T any] struct {
	fillers  FilterChain[T]
	handlers HandlersChain[T]
}

type EventHandler[T Eventer] struct {
	emitter driver.Emitter
	Composer[T]
	handlerEnds []HandlerEnd[T]
}

func (h *EventHandler[T]) consume(ctx context.Context, event types.Event) error {
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
			nsxctx := NewContext(ctx, h.emitter, event.SelfID, event.Time, msg, event.Replyer)
			nsxctx.handlers = handlerEnd.handlers
			nsxctx.Next()
		}()
	}
	return nil
}

type consumer interface {
	consume(ctx context.Context, event types.Event) error
}

func OnEvent[T Eventer](engine *Engine) *EventHandler[T] {
	eventHandler := &EventHandler[T]{
		emitter: engine.emitter,
	}
	eventHandler.root = eventHandler

	var eventer T
	engine.consumers[eventer.Type()] = eventHandler
	return eventHandler
}

type Engine struct {
	listener    driver.Listener
	emitter     driver.Emitter
	taskLen     int
	consumerNum int
	consumers   map[types.EventType]consumer
}

func Default(driver driver.Driver) *Engine {
	return &Engine{
		listener:    driver,
		emitter:     driver,
		taskLen:     10,
		consumerNum: runtime.NumCPU(),
		consumers:   make(map[types.EventType]consumer),
	}
}

func New(listener driver.Listener, emitter driver.Emitter) *Engine {
	return &Engine{
		listener:    listener,
		emitter:     emitter,
		taskLen:     10,
		consumerNum: runtime.NumCPU(),
		consumers:   make(map[types.EventType]consumer),
	}
}

func (e *Engine) SetTaskLen(taskLen int) {
	e.taskLen = taskLen
}

func (e *Engine) SetConsumerNum(consumerNum int) {
	e.consumerNum = consumerNum
}

func (e *Engine) Run(ctx context.Context) {
	task := make(chan types.Event, e.taskLen)
	for range e.consumerNum {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case event := <-task:
					for _, Type := range event.Types {
						if consumer, ok := e.consumers[Type]; ok {
							if err := consumer.consume(ctx, event); err != nil {
								slog.Error("Consume error", "error", err)
							}
						}
					}
				}
			}
		}()
	}
	if err := e.listener.Listen(ctx, task); err != nil {
		panic(err)
	}
}
