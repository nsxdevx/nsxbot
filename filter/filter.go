package filter

import (
	"slices"
	"strings"

	"github.com/atopos31/nsxbot/types"
)

type Filter[T any] func(data T) bool

func OnlyGroups(groups ...int64) Filter[types.EventGrMsg] {
	return func(data types.EventGrMsg) bool {
		return slices.Contains(groups, data.GroupId)
	}
}

func OnlyAtUsers(userIds ...string) Filter[types.EventGrMsg] {
	return func(data types.EventGrMsg) bool {
		ats, n := data.Ats()
		if n == 0 {
			return false
		}
		for _, at := range ats {
			if slices.Contains(userIds, at.QQ) {
				return true
			}
		}
		return false
	}
}

func OnlyGrUsers(users ...int64) Filter[types.EventGrMsg] {
	return func(data types.EventGrMsg) bool {
		return slices.Contains(users, data.UserId)
	}
}

func OnlyUsers(users ...int64) Filter[types.EventPvtMsg] {
	return func(data types.EventPvtMsg) bool {
		return slices.Contains(users, data.UserId)
	}
}

func OnCommand[T types.EventMsg](prefix string, commands ...string) Filter[T] {
	return func(msg T) bool {
		text, err := msg.TextFirst()
		if err != nil {
			return false
		}
		trimed := strings.TrimSpace(text.Text)
		if !strings.HasPrefix(trimed, prefix) {
			return false
		}
		parts := strings.Fields(trimed)
		if len(parts) == 0 {
			return false
		}
		cmd := strings.TrimPrefix(parts[0], prefix)
		for _, command := range commands {
			if strings.EqualFold(cmd, command) {
				return true
			}
		}
		return false
	}
}
