package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/atopos31/nsxbot"
	"github.com/atopos31/nsxbot/driver"
	"github.com/atopos31/nsxbot/types"
)

func main() {
	emitter := driver.NewHttpEmitter("http://localhost:4000")
	listener := driver.NewHttpListener(":8080")
	httpdriver := driver.NewHttpDriver(listener, emitter)
	bot := nsxbot.Default(httpdriver)

	pvt := nsxbot.OnEvent[types.EventPvtMsg](bot)
	pvt.Use(func(ctx *nsxbot.Context[types.EventPvtMsg]) {
		start := time.Now()
		ctx.Next()
		slog.Info("Process ", "time", time.Since(start))
	})

	pvt.Handle(func(ctx *nsxbot.Context[types.EventPvtMsg]) {
		text, err := ctx.Msg.TextFirst()
		if err != nil {
			slog.Error("Error parsing message", "error", err)
			return
		}
		slog.Info("Private Message", "message", text.Text)
		ctx.Reply(text.Text)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Run
	bot.Run(ctx)
}
