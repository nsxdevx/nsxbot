package main

import (
	"context"
	"log/slog"

	"github.com/nsxdevx/nsxbot"
	"github.com/nsxdevx/nsxbot/driver"
	"github.com/nsxdevx/nsxbot/event"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot := nsxbot.Default(driver.NewDriverHttp(":8080", "http://localhost:4000"))

	all := nsxbot.OnEvent[event.AllMessage](bot)
	all.Handle(func(ctx *nsxbot.Context[event.AllMessage]) {
		text, err := ctx.Msg.TextFirst()
		if err != nil {
			slog.Error("Error parsing message", "error", err)
			return
		}
		slog.Info("All Message", "message", text.Text)
	})

	pvt := nsxbot.OnEvent[event.PrivateMessage](bot)
	pvt.Handle(func(ctx *nsxbot.Context[event.PrivateMessage]) {
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
