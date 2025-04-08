package main

import (
	"context"
	"log/slog"

	"github.com/atopos31/nsxbot"
	"github.com/atopos31/nsxbot/driver"
	"github.com/atopos31/nsxbot/types"
)

func main() {
	emitter := driver.NewHttpEmitter("http://localhost:4000")
	listener := driver.NewHttpListener(":8080")
	httpdriver := driver.NewHttpDriver(listener, emitter)
	bot := nsxbot.Default(httpdriver)

	all := nsxbot.OnEvent[types.EventAllMsg](bot)
	all.Handle(func(ctx *nsxbot.Context[types.EventAllMsg]) {
		text, err := ctx.Msg.TextFirst()
		if err != nil {
			slog.Error("Error parsing message", "error", err)
			return
		}
		slog.Info("All Message", "message", text.Text)
	})

	pvt := nsxbot.OnEvent[types.EventPvtMsg](bot)
	pvt.Handle(func(ctx *nsxbot.Context[types.EventPvtMsg]) {
		text, err := ctx.Msg.TextFirst()
		if err != nil {
			slog.Error("Error parsing message", "error", err)
			return
		}
		slog.Info("Private Message", "message", text.Text)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Run
	bot.Run(ctx)
}
