package filter

import (
	"slices"

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
