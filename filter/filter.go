package filter

import (
	"slices"

	"github.com/atopos31/nsxbot/types"
)

type Filter[T any] func(data T) bool

func Group(data types.EventMessage) bool {
	return data.MessageType == "group"
}

func Private(data types.EventMessage) bool {
	return data.MessageType == "private"
}

func OnlyGroups(groups ...int64) Filter[types.EventMessage] {
	return func(data types.EventMessage) bool {
		return slices.Contains(groups, data.GroupID)
	}
}
