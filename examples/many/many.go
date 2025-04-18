package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	"github.com/atopos31/nsxbot"
	"github.com/atopos31/nsxbot/driver"
	"github.com/atopos31/nsxbot/filter"
	"github.com/atopos31/nsxbot/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	aili0uin, _ := strconv.ParseInt(os.Getenv("AILI_UIN_0"), 10, 64)
	aili1uin, _ := strconv.ParseInt(os.Getenv("AILI_UIN_1"), 10, 64)
	adminuin, _ := strconv.ParseInt(os.Getenv("ADMIN_UIN"), 10, 64)
	groupId, _ := strconv.ParseInt(os.Getenv("TEST_GROUP"), 10, 64)

	bot := nsxbot.Default(driver.NewWSverver(":8081", "/"))

	pvt := nsxbot.OnSelfsEvent[types.EventGrMsg](bot, aili0uin, aili1uin)

	pvt.Handle(func(ctx *nsxbot.Context[types.EventGrMsg]) {
		info, err := ctx.GetLoginInfo(ctx)
		if err != nil {
			slog.Error("Error getting login info", "error", err)
			return
		}
		slog.Info("ping!")
		var msg types.MeaasgeChain
		ctx.SendGrMsg(ctx, groupId, msg.Text("在!这里是:"+info.NickName))
	}, filter.OnlyGroups(groupId), filter.OnlyGrUsers(adminuin))

	// Run
	bot.Run(ctx)
}
