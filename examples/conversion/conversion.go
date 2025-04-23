package main

import (
	"context"
	"slices"
	"strings"

	nsx "github.com/nsxdevx/nsxbot"
	"github.com/nsxdevx/nsxbot/driver"
	"github.com/nsxdevx/nsxbot/filter"
	"github.com/nsxdevx/nsxbot/types"
)

func main() {
	driver := driver.NewDriverHttp(":8080", "http://localhost:4000")

	bot := nsx.Default(driver)

	pvt := nsx.OnEvent[types.EventGrMsg](bot)

	pvt.Handle(nsx.NewConversation(handler), filter.OnlyGroups(517170497))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Run
	bot.Run(ctx)
}

func handler(ctx0 *nsx.Context[types.EventGrMsg], sation *nsx.Sation[types.EventGrMsg]) {
	msg := ctx0.Msg
	text, err := msg.TextFirst()
	if err != nil {
		ctx0.Log.Error("Error parsing message", "error", err)
		return
	}
	cmd, err := text.Cmd("/")
	if !strings.EqualFold("set", cmd) || err != nil {
		msg.Reply(ctx0, "使用/set 开始设置！")
		return
	}
	var msgchain types.MeaasgeChain
	ctx0.SendGrMsg(ctx0, msg.GroupId, msgchain.Text("请选择:").Br().Text("1:test1").Br().Text("2:test2"))
	//等待下一条消息
	ctx1, err := sation.Await(ctx0)
	if err != nil {
		ctx0.Log.Error("Error parsing message", "error", err)
		return
	}
	msg1 := ctx1.Msg
	test1, err := msg1.TextFirst()
	if err != nil {
		ctx1.Log.Error("Error parsing message", "error", err)
		return
	}
	if !slices.Contains([]string{"1", "2"}, test1.Text) {
		msg1.Reply(ctx1, "请选择正确的选项！")
		return
	}
	msg1.Reply(ctx1, "选择"+test1.Text+"成功")
}
