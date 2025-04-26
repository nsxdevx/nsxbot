package main

import (
	"context"
	"os"
	"strconv"

	"github.com/nsxdevx/nsxbot"
	"github.com/nsxdevx/nsxbot/driver"
	"github.com/nsxdevx/nsxbot/filter"
	"github.com/nsxdevx/nsxbot/types"
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
			ctx.Log.Error("Error getting login info", "error", err)
			return
		}
		ctx.Log.Info("ping!")
		var msg types.MeaasgeChain
		ctx.SendGrMsg(ctx, groupId, msg.Text("在!这里是:"+info.NickName))
	}, filter.OnlyGroups(groupId), filter.OnlyGrUsers(adminuin))
	// Run
	bot.Run(ctx)
}
