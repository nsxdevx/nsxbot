package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	"github.com/atopos31/nsxbot"
	"github.com/atopos31/nsxbot/driver"
	"github.com/atopos31/nsxbot/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot := nsxbot.Default(driver.NewDriverHttp(":8080", "http://localhost:4000"))

	pvt := nsxbot.OnEvent[types.EventPvtMsg](bot)

	adminuin, _ := strconv.ParseInt(os.Getenv("ADMIN_UIN"), 10, 64)
	pvt.Handle(func(ctx *nsxbot.Context[types.EventPvtMsg]) {
		text, err := ctx.Msg.TextFirst()
		if err != nil {
			slog.Error("Error parsing message", "error", err)
			return
		}
		slog.Info("Private Message", "message", text.Text)
		ctx.Msg.Reply(ctx.Replayer, text.Text)
		var msg types.MeaasgeChain
		ctx.SendPvtMsg(ctx, adminuin, msg.Text("收到回复了吗？").Br().Text("2333333333"))
	})

	// Run
	bot.Run(ctx)
}
