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
	emitter := driver.NewHttpEmitter("http://localhost:4000")
	listener := driver.NewHttpListener(":8080")
	httpdriver := driver.NewHttpDriver(listener, emitter)
	bot := nsxbot.Default(httpdriver)

	gr := nsxbot.OnEvent[types.EventGrMsg](bot)

	gr1 := gr.Compose(filter.OnlyGroups(819085771))
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
	ge2.Handle(func(ctx *nsxbot.Context[types.EventGrMsg]) {
		text, err := ctx.Msg.TextFirst()
		if err != nil {
			slog.Error("Error parsing text message", "error", err)
			return
		}
		slog.Info("Group Message", "message", text.Text)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Run
	bot.Run(ctx)
}
