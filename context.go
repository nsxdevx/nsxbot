package nsxbot

import (
	"context"
	"log/slog"
	"math"

	"github.com/nsxdevx/nsxbot/driver"
	"github.com/nsxdevx/nsxbot/nlog"
	"github.com/nsxdevx/nsxbot/types"
)

const abortIndex int8 = math.MaxInt8 >> 1

type Context[T any] struct {
	context.Context
	driver.Emitter
	types.Replyer

	Time   int64
	SelfId int64
	Msg    T
	Log    *slog.Logger

	index    int8
	handlers HandlersChain[T]
}

func NewContext[T any](ctx context.Context, emitter driver.Emitter, selfId int64, time int64, data T, Replyer types.Replyer) Context[T] {
	return Context[T]{
		Context: ctx,
		Emitter: emitter,
		SelfId:  selfId,
		Time:    time,
		Msg:     data,
		Log:     nlog.Logger(),
		Replyer: Replyer,
		index:   -1,
	}
}

func (c *Context[T]) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		if c.handlers[c.index] != nil {
			c.handlers[c.index](c)
		}
		c.index++
	}
}

func (c *Context[T]) Abort() {
	c.index = abortIndex
}
