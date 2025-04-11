package main

import (
	"context"
	"log/slog"

	"github.com/atopos31/nsxbot"
	"github.com/atopos31/nsxbot/driver"
	"github.com/atopos31/nsxbot/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot := nsxbot.Default(ctx, driver.NewDriverHttp(":8080", "http://localhost:4000"))

	pvt := nsxbot.OnEvent[types.EventPvtMsg](bot)

	pvt.Handle(func(ctx *nsxbot.Context[types.EventPvtMsg]) {
		text, err := ctx.Msg.TextFirst()
		if err != nil {
			slog.Error("Error parsing message", "error", err)
			return
		}
		slog.Info("Private Message", "message", text.Text)
		ctx.Reply(text.Text)
		var msg types.MeaasgeChain
		ctx.SendPvtMsg(ctx, 2945294768, msg.Text("收到回复了吗？").Br().Text("2333333333"))
	})

	// Run
	bot.Run(ctx)
}
