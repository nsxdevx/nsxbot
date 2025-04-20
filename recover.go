package nsxbot

func Recovery[T any]() HandlerFunc[T] {
	return func(ctx *Context[T]) {
		defer func() {
			if err := recover(); err != nil {
				ctx.Log.Error("Handler Panic", "err", err, "time", ctx.Time, "selfId", ctx.SelfId)
			}
		}()
		ctx.Next()
	}
}
