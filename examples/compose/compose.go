package main

import (
	"context"
	"log/slog"

	"github.com/nsxdevx/nsxbot"
	"github.com/nsxdevx/nsxbot/driver"
	"github.com/nsxdevx/nsxbot/filter"
	"github.com/nsxdevx/nsxbot/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot := nsxbot.Default(driver.NewDriverHttp(":8080", "http://localhost:4000"))

	gr := nsxbot.OnEvent[types.EventGrMsg](bot)

	gr1 := gr.Compose(filter.OnlyGroups(819085771), filter.OnlyAtUsers("123456789"))
	gr1.Handle(func(ctx *nsxbot.Context[types.EventGrMsg]) {
		text, err := ctx.Msg.TextFirst()
		if err != nil {
			slog.Error("Error parsing text message", "error", err)
			return
		}
		slog.Info("Group Message", "message", text.Text)
	})

	gr1.Handle(func(ctx *nsxbot.Context[types.EventGrMsg]) {
		face, err := ctx.Msg.FaceFirst()
		if err != nil {
			slog.Error("Error parsing face message", "error", err)
			return
		}
		slog.Info("Group Message", "face", face.Id)
	})

	ge2 := gr.Compose(filter.OnlyGroups(517170497))
	ge2.Handle(ontext)

	// Run
	bot.Run(ctx)
}

func ontext(ctx *nsxbot.Context[types.EventGrMsg]) {
	text, err := ctx.Msg.TextFirst()
	if err != nil {
		slog.Error("Error parsing text message", "error", err)
		return
	}
	slog.Info("Group Message", "message", text.Text)
}
