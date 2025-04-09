package main

import (
	"context"
	"log/slog"

	"github.com/atopos31/nsxbot"
	"github.com/atopos31/nsxbot/driver"
	"github.com/atopos31/nsxbot/filter"
	"github.com/atopos31/nsxbot/types"
)

func main() {
	bot := nsxbot.Default(driver.NewDriverHttp(":8080", "http://localhost:4000"))

	gr := nsxbot.OnEvent[types.EventGrMsg](bot)
	gr.Handle(func(ctx *nsxbot.Context[types.EventGrMsg]) {
		text, err := ctx.Msg.TextFirst()
		if err != nil {
			slog.Error("Error parsing message", "error", err)
			return
		}
		slog.Info("Group Message", "message", text.Text)
	}, filter.OnlyGroups(819085771))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Run
	bot.Run(ctx)
}
