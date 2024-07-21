package core

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

type Event any
type EventPack []Event

type AggregateState interface {
	Apply(event Event) AggregateState
}

type Aggregate[TState AggregateState] struct {
	id     string
	events []Event
	state  TState
	err    error
}

func (a *Aggregate[T]) ProcessCommand(handler func(*T, EventRaiser) error) (EventPack, error) {
	if a.err != nil {
		panic("aggregagte is in error state")
	}
	events := EventPack(make([]Event, 0))
	apply := applier{func(event Event) {
		a.apply(event)
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

func (a *Aggregate[TState]) apply(event Event) {
	a.state.Apply(event)
}

func (a *Aggregate[TState]) Id() string {
	return a.id
}

func (a *Aggregate[TState]) GetState() (AggregateState, EventPack) {
	return a.state, a.events
}

func (a *Aggregate[TState]) Restore(state AggregateState) {
	a.state = state.(TState)
}

func (a *Aggregate[TState]) StateType() reflect.Type {
	return reflect.TypeOf(a.state)
}

type IAggregate interface {
	Id() string
	GetState() (AggregateState, EventPack)
	Restore(AggregateState)
	StateType() reflect.Type
}

type Persister interface {
	Store(any, EventPack) error
	Restore(id string) (any, error)
}

type JosnbPersister[T any] struct {
}

func (j JosnbPersister[T]) Store(state any, pack EventPack) error {
	fmt.Printf("state: [%v]\n", state)
	fmt.Printf("events: [%v]\n", pack)
	return nil
}

func (j JosnbPersister[T]) Restore(id string) (any, error) {
	s := `{"name" : "123"}`
	res := new(T)
	err := json.Unmarshal([]byte(s), res)
	if err != nil {
		return *new(T), err
	}
	return *res, nil
}

func (j JosnbPersister[T]) SetTransaction(tx string) {
}

type Repository interface {
	Load(ctx context.Context, id string, aggregate IAggregate) error
	Save(ctx context.Context, aggregate IAggregate) error
}

type PostgresPersister interface {
	Persister
	SetTransaction(tx string)
}

func NewDummyRepository() *DummyRepository {
	persisters := map[reflect.Type]func() PostgresPersister{reflect.TypeOf((*TestState)(nil)).Elem(): func() PostgresPersister { return JosnbPersister[TestState]{} }}
	return &DummyRepository{persisters}
}

type DummyRepository struct {
	persisters map[reflect.Type]func() PostgresPersister
}

type ConcurrentScope struct {
	createRepo func() Repository
}

func NewConcurrentScope(createRepo func() Repository) *ConcurrentScope {
	return &ConcurrentScope{createRepo}
}

func (c ConcurrentScope) Run(runFunc func(repo Repository) error) error {
	repo := c.createRepo()
	return runFunc(repo)
}

func (p DummyRepository) Load(ctx context.Context, id string, aggregate IAggregate) error {
	persister, ok := p.persisters[aggregate.StateType()]
	if !ok {
		return fmt.Errorf("persister not found for [%s]", aggregate.StateType().Name())
	}
	state, err := persister().Restore(id)
	if err != nil {
		return err
	}
	aggregate.Restore(state.(AggregateState))
	return nil
}

func (p DummyRepository) Save(ctx context.Context, aggregate IAggregate) error {
	return nil
}

type TestState struct {
	Name string
}

func (t TestState) Apply(event Event) AggregateState {
	return TestState{t.Name + "abc"}
}

type TestAgg struct {
	Aggregate[TestState]
}
