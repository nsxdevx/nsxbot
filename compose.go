package nsxbot

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/atopos31/nsxbot/filter"
	"github.com/atopos31/nsxbot/types"
)

type FilterChain[T any] []filter.Filter[T]

func (f FilterChain[T]) debug() string {
	var info string
	for _, filter := range f {
		info += strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(filter).Pointer()).Name()+"->", "main.main.")
	}
	return info
}

type HandlersChain[T any] []HandlerFunc[T]

type HandlerFunc[T any] func(ctx *Context[T])

type Composer[T types.Eventer] struct {
	handlers HandlersChain[T]
	filters  FilterChain[T]
	root     *EventHandler[T]
}

// Use adds handlers to the Composer.
func (c *Composer[T]) Use(handlers ...HandlerFunc[T]) {
	c.handlers = append(c.handlers, handlers...)
}

// Filit adds filters to the Composer.
func (c *Composer[T]) Filit(fillers ...filter.Filter[T]) {
	c.filters = append(c.filters, fillers...)
}

// Compose creates a new Composer with the given filters.
func (c *Composer[T]) Compose(fillers ...filter.Filter[T]) *Composer[T] {
	return &Composer[T]{
		handlers: c.handlers,
		root:     c.root,
		filters:  c.combineFilters(fillers),
	}
}

// Handle adds a handler to the Composer.
func (c *Composer[T]) Handle(handler HandlerFunc[T], filters ...filter.Filter[T]) {
	handlerEnd := HandlerEnd[T]{
		fillers:  c.combineFilters(filters),
		handlers: append(c.handlers, handler),
	}
	c.root.handlerEnds = append(c.root.handlerEnds, handlerEnd)
}

func (c *Composer[T]) combineFilters(filters FilterChain[T]) FilterChain[T] {
	finalSize := len(c.filters) + len(filters)
	mergedFilters := make(FilterChain[T], finalSize)
	copy(mergedFilters, c.filters)
	copy(mergedFilters[len(c.filters):], filters)
	return mergedFilters
}
