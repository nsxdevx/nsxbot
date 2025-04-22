package main

import (
	"context"
	"log/slog"

	"github.com/nsxdevx/nsxbot"
	"github.com/nsxdevx/nsxbot/driver"
	"github.com/nsxdevx/nsxbot/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot := nsxbot.Default(driver.NewDriverHttp(":8080", "http://localhost:4000"))

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

	// Run
	bot.Run(ctx)
}
