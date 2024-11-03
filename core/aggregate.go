package core

import (
	"errors"
	"fmt"
)

var ErrAggregateHasError = errors.New("aggregate has error")

type Aggregate[TState AggregateState] struct {
	id      AggregateId
	version Version
	events  EventPack
	state   TState
	err     error
}

func (a *Aggregate[TState]) checkError() {
	if a.err != nil {
		panic(ErrAggregateHasError)
	}
}

func (a *Aggregate[TState]) ProcessCommand(handler func(*TState, EventRaiser) error) (EventPack, error) {
	a.checkError()
	events := EventPack(make([]Event, 0))
	apply := applier{func(event Event) {
		a.state = a.state.Apply(event).(TState)
		events = append(events, event)
	}}
	err := handler(&a.state, apply)
	if err == nil {
		return events, nil
	}
	// If an error occur after command has spawned any events
	// then the  aggregate is considered in corrupted state and can't be used anymore
	if IsEmpty(events) {
		return nil, err
	}
	a.err = err
	a.events = nil
	return nil, err
}

func (a *Aggregate[TState]) Id() AggregateId {
	a.checkError()
	return a.id
}

func (a *Aggregate[T]) State() T {
	a.checkError()
	return a.state
}

func PanicUnsupportedEvent(event Event) error {
	panic(fmt.Sprintf("unsupported event %T", event))
}

func (a *Aggregate[T]) Initialize(id AggregateId, created Event) {
	if a.version > 0 {
		panic(fmt.Errorf("aggregate is already initialized"))
	}
	a.id = id
	a.version = 0
	var empty T
	a.state = empty
	a.state = a.state.Apply(created).(T)
	a.events = []Event{created}
	a.err = nil
}

func (a *Aggregate[TState]) Store(persistFunc func(AggregateState, EventPack, Version) error) error {
	a.checkError()
	err := persistFunc(a.state, a.events, a.version)
	if err != nil {
		return err
	}
	a.events = nil
	a.version++
	return nil
}

func (a *Aggregate[TState]) Restore(id AggregateId, state AggregateState, version Version) {
	a.id = id
	a.state = state.(TState)
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

type IAggregate interface {
	Id() AggregateId
	Version() Version
	Error() error
	Store(persistFunc func(AggregateState, EventPack, Version) error) error
	Restore(id AggregateId, state AggregateState, version Version)
}
