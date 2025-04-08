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

	gr := nsxbot.OnEvent[types.EventPvtMsg](bot)
	gr.Handle(func(ctx *nsxbot.Context[types.EventPvtMsg]) {
		text, err := ctx.Msg.TextFirst()
		if err != nil {
			slog.Error("Error parsing message", "error", err)
			return
		}
		cmd, err := text.Cmd("/")
		if err != nil {
			slog.Error("Error parsing command", "error", err)
			return
		} else {
			slog.Info("Command", "command", cmd)
		}
		arg, err := text.CmdIndex("/", 0)
		if err != nil {
			slog.Error("Error parsing command index", "error", err)
		} else {
			slog.Info("Command index", "arg", arg)
		}
		value, err := text.CmdKey("key")
		if err != nil {
			slog.Error("Error parsing command key", "error", err)
		} else {
			slog.Info("Command key", "value", value)
		}

	}, filter.OnCommand[types.EventPvtMsg]("/", "echo"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Run
	bot.Run(ctx)
}
