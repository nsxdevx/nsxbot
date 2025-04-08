package nsxbot

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/atopos31/nsxbot/driver"
	"github.com/atopos31/nsxbot/types"
)

type HandlerEnd[T any] struct {
	fillers  FilterChain[T]
	handlers HandlersChain[T]
}

type EventHandler[T Eventer] struct {
	emitter driver.Emitter
	Composer[T]
	handlerEnds []HandlerEnd[T]
}

type Eventer interface {
	Type() string
}

func (h *EventHandler[T]) consume(ctx context.Context, event types.Event) error {
	var msg T
	if err := json.Unmarshal(event.RawData, &msg); err != nil {
		return err
	}
	event.RawData = nil
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

type Consumer interface {
	consume(ctx context.Context, event types.Event) error
}

type Config struct {
	taskSize    int
	consumerNum int
}

type Engine struct {
	driver      driver.Driver
	task        chan types.Event
	consumerNum int
	consumers   map[types.EventType]Consumer
}

func Default(driver driver.Driver) *Engine {
	return &Engine{
		driver:      driver,
		task:        make(chan types.Event, 10),
		consumerNum: 5,
		consumers:   make(map[types.EventType]Consumer, 5),
	}
}

func New(driver driver.Driver, config Config) *Engine {
	return &Engine{
		driver:      driver,
		task:        make(chan types.Event, config.taskSize),
		consumerNum: config.consumerNum,
		consumers:   make(map[types.EventType]Consumer, config.consumerNum),
	}
}

func OnEvent[T Eventer](engine *Engine) *EventHandler[T] {
	eventHandler := &EventHandler[T]{
		emitter: engine.driver,
	}
	eventHandler.root = eventHandler

	var eventer T
	engine.consumers[eventer.Type()] = eventHandler
	return eventHandler
}

func (e *Engine) Emitter() driver.Emitter {
	return e.driver
}

func (e *Engine) Run(ctx context.Context) error {
	for range e.consumerNum {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case event := <-e.task:
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
	if err := e.driver.Listen(ctx, e.task); err != nil {
		slog.Error("Listen error", "error", err)
		return err
	}
	return nil
}
