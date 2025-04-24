<div align="center">

# NsxBot OneBot Framework

![nsxbot](https://socialify.git.ci/atopos31/nsxbot/image?font=Inter&language=1&logo=https%3A%2F%2Fonebot.dev%2Flogo.png&name=1&owner=1&pattern=Circuit+Board&stargazers=1&theme=Auto)

[![Go](https://img.shields.io/badge/Go-00ADD8.svg?logo=go&logoColor=white)](https://go.dev/)
[![Badge](https://img.shields.io/badge/OneBot-11-black)](https://github.com/botuniverse/onebot-11)
[![License](https://img.shields.io/badge/License-unlicense-green)](https://github.com/nsxdevx/nsxbot/blob/master/LICENSE)
[![qq group](https://img.shields.io/badge/Group-881412730-red?style=flat-square&logo=tencent-qq)](https://qm.qq.com/cgi-bin/qm/qr?k=d5DcTIKBYVmaHZHZ4BqwKaXop4ePjrh_&jump_from=webapi&authKey=nY7Yhr6GhgbS28XBw0nrH4M3tutmPF9U1+5m7GCaRgaABTqBHkTcHC1l1Sa1NFrh)

</div>

## 简介

NsxBot 是一个使用 [Go](https://go.dev/) 语言编写，基于 [OneBot 11](https://github.com/botuniverse/onebot-11) 协议的聊天机器人框架。

提供类似Web框架风格的API，如果你是一个Go Web开发者，那么你可以非常方便的使用Nsxbot。

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
> [!IMPORTANT] 
> 未发布第一版测试，你会拉取到仓库的最新提交，不保证可靠，框架正在开发中......
### 运行
回复示例：
```go
package main

import (
	"context"
	"os"
	"strconv"

	"github.com/nsxdevx/nsxbot"
	"github.com/nsxdevx/nsxbot/driver"
	"github.com/nsxdevx/nsxbot/types"
)

func main() {
	driver := driver.NewDriverHttp(":8080", "http://localhost:4000")

	bot := nsxbot.Default(driver)

	pvt := nsxbot.OnEvent[types.EventPvtMsg](bot)

	adminuin, _ := strconv.ParseInt(os.Getenv("ADMIN_UIN"), 10, 64)
	pvt.Handle(func(ctx *nsxbot.Context[types.EventPvtMsg]) {
		msg := ctx.Msg
		text, err := msg.TextFirst()
		if err != nil {
			ctx.Log.Error("Error parsing message", "error", err)
			return
		}
		ctx.Log.Info("Private Message", "message", text.Text)
		msg.Reply(ctx, text.Text)
		var msgchain types.MeaasgeChain
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