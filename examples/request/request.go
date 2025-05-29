package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	"github.com/nsxdevx/nsxbot"
	"github.com/nsxdevx/nsxbot/driver"
	"github.com/nsxdevx/nsxbot/event"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot := nsxbot.Default(driver.NewDriverHttp(":8080", "http://localhost:4000"))

	req := nsxbot.OnEvent[event.GroupRequest](bot)

	adminuin, _ := strconv.ParseInt(os.Getenv("ADMIN_UIN"), 10, 64)
	req.Handle(func(ctx *nsxbot.Context[event.GroupRequest]) {
		if ctx.Msg.UserId == adminuin {
			slog.Info("Friend Request", "user", ctx.Msg.UserId, "comment", ctx.Msg.Comment)
			ctx.Msg.Reply(ctx.Replyer, true, "admin")
		}
	})

	greq := nsxbot.OnEvent[event.GroupRequest](bot)
	greq.Handle(func(ctx *nsxbot.Context[event.GroupRequest]) {
		slog.Info("Group Request", "user", ctx.Msg.UserId, "comment", ctx.Msg.Comment)
		ctx.Msg.Reply(ctx.Replyer, false, "不要")
	})

	// Run
	bot.Run(ctx)
}
