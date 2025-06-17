package main

import (
	"context"
	"os"
	"strconv"

	"github.com/nsxdevx/nsxbot"
	"github.com/nsxdevx/nsxbot/driver"
	"github.com/nsxdevx/nsxbot/event"
	"github.com/nsxdevx/nsxbot/schema"
)

func main() {
	driver := driver.NewDriverHttp(":8080", "http://localhost:4000")

	bot := nsxbot.Default(driver)

	pvt := nsxbot.OnEvent[event.PrivateMessage](bot)

	adminuin, _ := strconv.ParseInt(os.Getenv("ADMIN_UIN"), 10, 64)
	pvt.Handle(func(ctx *nsxbot.Context[event.PrivateMessage]) {
		msg := ctx.Msg
		text, err := msg.TextFirst()
		if err != nil {
			ctx.Log.Error("Error parsing message", "error", err)
			return
		}
		ctx.Log.Info("Private Message", "message", text.Text)
		msg.Reply(ctx, text.Text)
		var msgchain schema.MessageChain
		ctx.SendPvtMsg(ctx, adminuin, msgchain.Text("收到回复了吗？").Br().Face("4"))
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Run
	bot.Run(ctx)
}
