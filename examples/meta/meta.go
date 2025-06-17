package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	"github.com/nsxdevx/nsxbot"
	"github.com/nsxdevx/nsxbot/driver"
	"github.com/nsxdevx/nsxbot/event"
	"github.com/nsxdevx/nsxbot/schema"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot := nsxbot.Default(driver.NewWSverver(":8081", "/"))

	pvt := nsxbot.OnEvent[event.PrivateMessage](bot)

	adminuin, _ := strconv.ParseInt(os.Getenv("ADMIN_UIN"), 10, 64)
	pvt.Handle(func(ctx *nsxbot.Context[event.PrivateMessage]) {
		text, err := ctx.Msg.TextFirst()
		if err != nil {
			slog.Error("Error parsing message", "error", err)
			return
		}
		slog.Info("Private Message", "message", text.Text)
		var msg schema.MessageChain
		ctx.SendPvtMsg(ctx, adminuin, msg.Text("收到回复了吗？").Br().Text("2333333333"))
	})

	life := nsxbot.OnEvent[event.LifeMeta](bot)

	life.Handle(func(ctx *nsxbot.Context[event.LifeMeta]) {
		slog.Info("Life Meta", "meta", ctx.Msg.SubType)
	})

	heart := nsxbot.OnEvent[event.HeartMeta](bot)
	heart.Handle(func(ctx *nsxbot.Context[event.HeartMeta]) {
		slog.Info("Heart Meta", "meta", ctx.Msg.Status)
	})

	// Run
	bot.Run(ctx)
}
