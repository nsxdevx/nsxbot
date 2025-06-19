<div align="center">

# NsxBot OneBot Framework

![nsxbot](https://socialify.git.ci/atopos31/nsxbot/image?font=Inter&language=1&logo=https%3A%2F%2Fonebot.dev%2Flogo.png&name=1&owner=1&pattern=Circuit+Board&stargazers=1&theme=Auto)

[![Go Reference](https://pkg.go.dev/badge/github.com/nsxdevx/nsxbot.svg)](https://pkg.go.dev/github.com/nsxdevx/nsxbot)
[![Badge](https://img.shields.io/badge/OneBot-11-black)](https://github.com/botuniverse/onebot-11)
[![License](https://img.shields.io/badge/License-unlicense-green)](https://github.com/nsxdevx/nsxbot/blob/master/LICENSE)
[![qq group](https://img.shields.io/badge/Group-881412730-red?style=flat-square&logo=tencent-qq)](https://qm.qq.com/cgi-bin/qm/qr?k=d5DcTIKBYVmaHZHZ4BqwKaXop4ePjrh_&jump_from=webapi&authKey=nY7Yhr6GhgbS28XBw0nrH4M3tutmPF9U1+5m7GCaRgaABTqBHkTcHC1l1Sa1NFrh)

</div>

> **⚠️ 注意：** 本项目目前处于 v0.x.y 阶段，API 尚不稳定，随时可能发生变更。请勿大规模使用，或锁定到具体的 commit/tag。

## 简介

NsxBot 是一个使用 [Go](https://go.dev/) 语言编写，基于 [OneBot 11](https://github.com/botuniverse/onebot-11) 协议的聊天机器人框架。

提供类似Web框架风格的API，如果你是一个Go Web开发者，那么你可以非常方便的使用NsxBot。

## 特性
- http，websocket 协议支持
- 支持多客户端统一处理
- 泛型支持，远离any
- 中间件支持
- 过滤器支持
- 事件分组监听
- 自由组合与可扩展性

## 快速开始

### 获取

```sh
go get -u github.com/nsxdevx/nsxbot
```

### 运行
示例：
```go
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
```
## 参考
- [OneBot 11](https://github.com/botuniverse/onebot-11)
- [OneBot 大典](https://github.com/tanebijs/onebot-pedia)
- [NapCat 接口文档](https://napcat.apifox.cn/)