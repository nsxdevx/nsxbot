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
	bot := nsxbot.Default(ctx, driver.NewWSverver(":8081", "/"))

	pvt := nsxbot.OnEvent[types.EventPvtMsg](bot)

	adminuin, _ := strconv.ParseInt(os.Getenv("ADMIN_UIN"), 10, 64)
	pvt.Handle(func(ctx *nsxbot.Context[types.EventPvtMsg]) {
		text, err := ctx.Msg.TextFirst()
		if err != nil {
			slog.Error("Error parsing message", "error", err)
			return
		}
		slog.Info("Private Message", "message", text.Text)
		ctx.Reply(text.Text)
		var msg types.MeaasgeChain
		ctx.SendPvtMsg(ctx, adminuin, msg.Text("收到回复了吗？").Br().Text("2333333333"))
	})

	life := nsxbot.OnEvent[types.LifeMeta](bot)

	life.Handle(func(ctx *nsxbot.Context[types.LifeMeta]) {
		slog.Info("Life Meta", "meta", ctx.Msg.SubType)
	})

	heart := nsxbot.OnEvent[types.HeartMeta](bot)
	heart.Handle(func(ctx *nsxbot.Context[types.HeartMeta]) {
		slog.Info("Heart Meta", "meta", ctx.Msg.Status)
	})

	// Run
	bot.Run(ctx)
}
