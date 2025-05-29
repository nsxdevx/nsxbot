package nsxbot

import (
	"context"
	"sync"

	"github.com/nsxdevx/nsxbot/event"
	"github.com/nsxdevx/nsxbot/filter"
)

type SessionMsg interface {
	event.Messager
	SessionKey() string
}

type Sation[T event.Messager] struct {
	sessionChan chan *Context[T]
}

func (s *Sation[T]) Await(ctx context.Context, fillers ...filter.Filter[T]) (*Context[T], error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-s.sessionChan:
		return msg, nil
	}
}

type SessionStore[T event.Messager] struct {
	mu       sync.Mutex
	sessions map[string]*Sation[T]
}

func (s *SessionStore[T]) Set(key string, ctx *Context[T]) (*Sation[T], bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[key]; ok {
		s.sessions[key].sessionChan <- ctx
		return s.sessions[key], false
	} else {
		s.sessions[key] = &Sation[T]{
			sessionChan: make(chan *Context[T], 1),
		}
		return s.sessions[key], true
	}
}

func (s *SessionStore[T]) Del(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	close(s.sessions[key].sessionChan)
	delete(s.sessions, key)
}

type SessionHandler[T event.Messager] = func(ctx *Context[T], sation *Sation[T])

func NewConversation[T SessionMsg](handler SessionHandler[T]) HandlerFunc[T] {
	store := &SessionStore[T]{
		sessions: make(map[string]*Sation[T]),
	}
	return func(ctx *Context[T]) {
		key := ctx.Msg.SessionKey()
		sation, first := store.Set(key, ctx)
		if first {
			handler(ctx, sation)
			store.Del(key)
		}
	}
}
