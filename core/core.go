package core

import (
	"errors"
)

type Event any
type EventPack []Event
type AggregateId string
type Version uint64

type AggregateState interface {
	Apply(event Event) AggregateState
}

var (
	ErrNoEvents      = errors.New("no events")
	ErrTooManyEvents = errors.New("too many events")
)

func EventOfType[T any](pack EventPack) (T, error) {
	e := EventsOfType[T](pack)
	var evt T
	if len(e) == 0 {
		return evt, ErrNoEvents
	} else if len(e) > 1 {
		return evt, ErrTooManyEvents
	} else {
		return e[0], nil
	}
}

func EventsOfType[T any](pack EventPack) []T {
	res := make([]T, 0)
	for _, e := range pack {
		switch evt := e.(type) {
		case T:
			res = append(res, evt)
		}
	}
	return res
}

func IsEmpty(pack EventPack) bool {
	return len(pack) == 0
}
