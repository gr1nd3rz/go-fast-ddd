package core

type Aggregate[TState AggregateState] struct {
	id      AggregateId
	version Version
	events  []Event
	state   TState
	err     error
}

func (a *Aggregate[TState]) ProcessCommand(handler func(*TState, EventRaiser) error) (EventPack, error) {
	if a.err != nil {
		panic(ErrAggregateHasError)
	}
	events := EventPack(make([]Event, 0))
	apply := applier{func(event Event) {
		a.state.Apply(event)
		events = append(events, event)
	}}
	err := handler(&a.state, apply)
	if err == nil {
		return events, nil
	}
	a.err = err
	a.events = nil
	return nil, err
}

func (a *Aggregate[TState]) Id() AggregateId {
	return a.id
}

func (a *Aggregate[TState]) Store(persistFunc func(AggregateState, EventPack, Version) error) error {
	if a.err != nil {
		panic(ErrAggregateHasError)
	}
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

type IAggregate interface {
	Id() AggregateId
	Store(persistFunc func(AggregateState, EventPack, Version) error) error
	Restore(id AggregateId, state AggregateState, version Version)
}
