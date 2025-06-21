package main

import (
	"context"
	"log/slog"

	"github.com/nsxdevx/nsxbot"
	"github.com/nsxdevx/nsxbot/driver"
	"github.com/nsxdevx/nsxbot/event"
	"github.com/nsxdevx/nsxbot/filter"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot := nsxbot.Default(driver.NewDriverHttp(":8080", "http://localhost:4000"))

	gr := nsxbot.OnEvent[event.PrivateMessage](bot)
	gr.Handle(func(ctx *nsxbot.Context[event.PrivateMessage]) {
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
		value, err := text.CmdValue("key")
		if err != nil {
			slog.Error("Error parsing command key", "error", err)
		} else {
			slog.Info("Command key", "value", value)
		}

	}, filter.OnCommand[event.PrivateMessage]("/", "echo"))

	// Run
	bot.Run(ctx)
}
