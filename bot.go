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

type EventHandler[T any] struct {
	emitter driver.Emitter
	Composer[T]
	handlerEnds []HandlerEnd[T]
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
	consumers   map[types.PostType]Consumer
}

func Default(driver driver.Driver) *Engine {
	return &Engine{
		driver:      driver,
		task:        make(chan types.Event, 10),
		consumerNum: 5,
		consumers:   make(map[types.PostType]Consumer),
	}
}

func NewEngine(driver driver.Driver, config Config) *Engine {
	return &Engine{
		driver:      driver,
		task:        make(chan types.Event, config.taskSize),
		consumerNum: config.consumerNum,
	}
}

func (e *Engine) Emitter() driver.Emitter {
	return e.driver
}

func (e *Engine) OnMessage() *EventHandler[types.EventMessage] {
	if v, ok := e.consumers[types.POST_TYPE_MESSAGE]; ok {
		return v.(*EventHandler[types.EventMessage])
	}
	messageHandler := &EventHandler[types.EventMessage]{
		emitter: e.driver,
	}
	e.consumers[types.POST_TYPE_MESSAGE] = messageHandler
	messageHandler.root = messageHandler
	return messageHandler
}

func (e *Engine) OnNotice() *EventHandler[types.EventNotice] {
	if v, ok := e.consumers[types.POST_TYPE_NOTICE]; ok {
		return v.(*EventHandler[types.EventNotice])
	}
	noticeHandler := &EventHandler[types.EventNotice]{
		emitter: e.driver,
	}
	e.consumers[types.POST_TYPE_NOTICE] = noticeHandler
	noticeHandler.root = noticeHandler
	return noticeHandler
}

func (e *Engine) OnRequest() *EventHandler[types.EventRequest] {
	if v, ok := e.consumers[types.POST_TYPE_REQUEST]; ok {
		return v.(*EventHandler[types.EventRequest])
	}
	requestHandler := &EventHandler[types.EventRequest]{
		emitter: e.driver,
	}
	e.consumers[types.POST_TYPE_REQUEST] = requestHandler
	requestHandler.root = requestHandler
	return requestHandler
}

func (e *Engine) OnMeta() *EventHandler[types.EventMeta] {
	if v, ok := e.consumers[types.POST_TYPE_META_ENEVT]; ok {
		return v.(*EventHandler[types.EventMeta])
	}
	metaEventHandler := &EventHandler[types.EventMeta]{
		emitter: e.driver,
	}
	e.consumers[types.POST_TYPE_META_ENEVT] = metaEventHandler
	metaEventHandler.root = metaEventHandler
	return metaEventHandler
}

func (e *Engine) Run(ctx context.Context) error {
	for range e.consumerNum {
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case event := <-e.task:
					if handler, ok := e.consumers[event.PostType]; ok {
						ctx := context.Background()
						if err := handler.consume(ctx, event); err != nil {
							slog.Error("Consume error", "error", err, "post_type", event.PostType)
						}
					}
				}
			}
		}(ctx)
	}
	if err := e.driver.Listen(ctx, e.task); err != nil {
		slog.Error("Listen error", "error", err)
		return err
	}
	return nil
}
