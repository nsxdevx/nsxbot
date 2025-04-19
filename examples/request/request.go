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

	req := nsxbot.OnEvent[types.EventFriendReq](bot)

	adminuin, _ := strconv.ParseInt(os.Getenv("ADMIN_UIN"), 10, 64)
	req.Handle(func(ctx *nsxbot.Context[types.EventFriendReq]) {
		if ctx.Msg.UserId == adminuin {
			slog.Info("Friend Request", "user", ctx.Msg.UserId, "comment", ctx.Msg.Comment)
			ctx.Msg.Reply(ctx.Replayer, true, "admin")
		}

	})

	// Run
	bot.Run(ctx)
}
