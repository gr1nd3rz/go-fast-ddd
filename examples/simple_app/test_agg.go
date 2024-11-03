package main

import (
	"errors"

	"github.com/gr1nd3rz/go-fast-ddd/core"
)

var _ core.AggregateState = testAggState{}

type nestedEntity struct {
	MyString string
	MySlice  []string
	MyMap    map[string]string
}

type testAggState struct {
	MyString string
	MySlice  []nestedEntity
	MyMap    map[string]nestedEntity
}

func newTestAgg(id core.AggregateId) testAgg {
	agg := testAgg{}
	agg.Initialize(id, Created{})
	return agg
}

type Created struct {
}

type ValueUpdated struct {
	value string
}

func (t testAggState) Apply(event core.Event) core.AggregateState {
	switch e := event.(type) {
	case Created:
		t.MySlice = make([]nestedEntity, 0)
		t.MyString = "created"
	case ValueUpdated:
		t.MyString = e.value
	default:
		core.PanicUnsupportedEvent(event)
	}
	return t
}

type testAgg struct {
	core.Aggregate[testAggState]
}

const (
	guardErrorValue = "guard_error"
)

func (t *testAgg) SingleEventCommand(val string) (core.EventPack, error) {
	return t.ProcessCommand(func(tas *testAggState, er core.EventRaiser) error {
		if val == guardErrorValue {
			return errors.New("guard error")
		}
		er.Raise(ValueUpdated{val})
		return nil
	})
}

func (t *testAgg) MultipleEventsCommand(val string) (core.EventPack, error) {
	return t.ProcessCommand(func(tas *testAggState, er core.EventRaiser) error {
		er.Raise(ValueUpdated{val})
		if val == guardErrorValue {
			return errors.New("guard error")
		}
		er.Raise(ValueUpdated{val})
		return nil
	})
}
