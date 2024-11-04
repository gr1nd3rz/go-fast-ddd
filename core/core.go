package core

import (
	"errors"
	"fmt"

	"github.com/google/go-cmp/cmp"
)

var (
	ErrNoEvents          = errors.New("no events")
	ErrTooManyEvents     = errors.New("too many events")
	ErrAggregateHasError = errors.New("aggregate has error")
)

type (
	Event     any
	EventPack []Event
	Id        string
	Version   uint64
)

type EventRiser interface {
	Raise(event Event)
	RaisePack(pack EventPack)
	RaiseNotEqual(first any, second any, event Event)
	RaiseTrue(predicate bool, event Event)
}

type EventApplier interface {
	Apply(event Event)
}

type Storer interface {
	Store(persistFunc func(id Id, state any, events EventPack, version Version) error) error
}

type Restorer interface {
	Restore(id Id, state any, version Version)
}

type riser[T any] struct {
	a *Aggregate[T]
}

func (r *riser[T]) Raise(event Event) {
	r.a.raise(event)
}

func (r *riser[T]) RaisePack(pack EventPack) {
	for _, e := range pack {
		r.Raise(e)
	}
}

func (r *riser[T]) RaiseNotEqual(first any, second any, event Event) {
	if !cmp.Equal(first, second) {
		r.Raise(event)
	}
}

func (r *riser[T]) RaiseTrue(predicate bool, event Event) {
	if predicate {
		r.Raise(event)
	}
}

type Aggregate[T any] struct {
	id      Id
	version Version
	events  EventPack
	state   T
	err     error
	riser   riser[T]
}

func (a *Aggregate[T]) checkError() {
	if a.err != nil {
		panic(ErrAggregateHasError)
	}
}

func (a *Aggregate[T]) raise(event Event) {
	any(&a.state).(EventApplier).Apply(event)
	a.events = append(a.events, event)
}

func (a *Aggregate[T]) ProcessCommand(handler func(state *T, er EventRiser) error) (EventPack, error) {
	a.checkError()
	eventsCount := len(a.events)
	a.riser = riser[T]{a}
	err := handler(&a.state, &a.riser)
	if err == nil {
		return a.events[eventsCount:], nil
	}
	// If an error occur after command has spawned any events
	// then the  aggregate is considered in corrupted state and can't be used anymore
	if eventsCount == len(a.events) {
		return nil, err
	}
	a.err = err
	a.events = nil
	return nil, err
}

func (a *Aggregate[T]) Id() Id {
	a.checkError()
	return a.id
}

func (a *Aggregate[T]) State() T {
	a.checkError()
	return a.state
}

func (a *Aggregate[T]) Initialize(id Id, created Event) {
	if a.version > 0 {
		panic(fmt.Errorf("aggregate is already initialized"))
	}
	a.id = id
	a.version = 0
	var state T
	a.state = state
	a.events = nil
	a.err = nil
	a.raise(created)
}

func (a *Aggregate[T]) Store(persistFunc func(Id, any, EventPack, Version) error) error {
	a.checkError()
	err := persistFunc(a.id, a.state, a.events, a.version)
	if err != nil {
		return err
	}
	a.events = nil
	a.version++
	return nil
}

func (a *Aggregate[TState]) Restore(id Id, state any, version Version) {
	a.id = id
	s := state.(TState)
	a.state = s
	a.version = version
	a.events = nil
	a.err = nil
}

func (a *Aggregate[TState]) Error() error {
	return a.err
}

func (a *Aggregate[TState]) Version() Version {
	a.checkError()
	return a.version
}

func PanicUnsupportedEvent(event Event) error {
	panic(fmt.Sprintf("unsupported event %T", event))
}

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
