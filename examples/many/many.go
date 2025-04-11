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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot := nsxbot.New(ctx,
		driver.NewListenerHttp(":8080"),
		driver.NewEmitterHttp("http://localhost:4000"),
		driver.NewEmitterHttp("http://localhost:4001"),
	)

	pvt := nsxbot.OnSelfsEvent[types.EventGrMsg](bot, 3808139675, 3958045985)

	pvt.Handle(func(ctx *nsxbot.Context[types.EventGrMsg]) {
		info, err := ctx.GetLoginInfo(ctx)
		if err != nil {
			slog.Error("Error getting login info", "error", err)
			return
		}
		slog.Info("ping!")
		ctx.Reply("在!这里是:" + info.NickName)
	}, filter.OnlyGroups(517170497), filter.OnlyGrUsers(2945294768))

	// Run
	bot.Run(ctx)
}
