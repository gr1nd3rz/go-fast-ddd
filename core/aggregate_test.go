package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

var _ AggregateState = testAggState{}

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

func newTestAgg(id AggregateId) testAgg {
	agg := testAgg{}
	agg.Initialize(id, Created{})
	return agg
}

type Created struct {
}

type ValueUpdated struct {
	value string
}

func (t testAggState) Apply(event Event) AggregateState {
	switch e := event.(type) {
	case Created:
		t.MySlice = make([]nestedEntity, 0)
		t.MyString = "created"
	case ValueUpdated:
		t.MyString = e.value
	default:
		PanicUnsupportedEvent(event)
	}
	return t
}

type testAgg struct {
	Aggregate[testAggState]
}

const (
	guardErrorValue = "guard_error"
)

func (t *testAgg) SingleEventCommand(val string) (EventPack, error) {
	return t.ProcessCommand(func(tas *testAggState, er EventRaiser) error {
		if val == guardErrorValue {
			return errors.New("guard error")
		}
		er.Raise(ValueUpdated{val})
		return nil
	})
}

func (t *testAgg) MultipleEventsCommand(val string) (EventPack, error) {
	return t.ProcessCommand(func(tas *testAggState, er EventRaiser) error {
		er.Raise(ValueUpdated{val})
		if val == guardErrorValue {
			return errors.New("guard error")
		}
		er.Raise(ValueUpdated{val})
		return nil
	})
}

func TestAggregate(t *testing.T) {
	t.Run(`Given aggreate with non-zero version
		And with non-empty state
		And events is not empty
		And error is not empty
		When Initialize is called
		Then reset state to after Created event is applied
		And events to Created
		And version to zero
		And error to nil`, func(t *testing.T) {
		agg := testAgg{}
		agg.id = "id"
		agg.state.MyMap = map[string]nestedEntity{"val": {}}
		agg.events = append(agg.events, ValueUpdated{"val"})
		agg.err = errors.New("error")

		newId := AggregateId("new-id")
		agg.Initialize(newId, Created{})

		require.Equal(t, Version(0), agg.Version())
		require.Equal(t, 1, len(agg.events))
		require.Equal(t, testAggState{MyString: "created", MySlice: make([]nestedEntity, 0)}, agg.state)
		require.Nil(t, agg.Error())
	})

	t.Run(`Given a newly created aggreagate
		When the allowed command is called
		Then event is produced
		And state is updated`, func(t *testing.T) {

		agg := newTestAgg("id")
		const val = "allowed_value"
		pack, err := agg.SingleEventCommand(val)
		require.NoError(t, err)
		evt, err := EventOfType[ValueUpdated](pack)
		require.NoError(t, err)
		require.Equal(t, ValueUpdated{val}, evt)
		require.Equal(t, val, agg.State().MyString)
	})

	t.Run(`Given a newly created aggreagate
		When command is called and guard statement fails
		And the command hasn't produced events before the failure
		Then the error is returned
		And the aggregate is still valid`, func(t *testing.T) {
		agg := newTestAgg("id")
		pack, err := agg.SingleEventCommand(guardErrorValue)
		require.Error(t, err)
		require.True(t, IsEmpty(pack))
		require.Nil(t, agg.Error())
	})

	t.Run(`Given a newly created aggreagate
		When command is called and guard statement fails
		And the command has produced event before the failure
		Then the error is returned
		And the aggregate is in invalid state`, func(t *testing.T) {
		agg := newTestAgg("id")
		pack, err := agg.MultipleEventsCommand(guardErrorValue)
		require.Error(t, err)
		require.True(t, IsEmpty(pack))
		require.NotNil(t, agg.Error())
	})

	t.Run(`Given a newly created aggregate
		When Store is called
		And persistFunc doesn't return an error
		Then should pass correct valued into persistFunc
		And increment version
		And cleanup events
	`, func(t *testing.T) {
		agg := newTestAgg("id")
		var pState AggregateState
		var pEventPack EventPack
		var pVersion Version
		err := agg.Store(func(as AggregateState, ep EventPack, v Version) error {
			pState = as
			pEventPack = ep
			pVersion = v
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, testAggState{MyString: "created", MySlice: make([]nestedEntity, 0)}, pState)
		require.Equal(t, EventPack{Created{}}, pEventPack)
		require.Equal(t, Version(0), pVersion)
		require.Empty(t, agg.events)
		require.Equal(t, Version(1), agg.Version())
		require.Equal(t, testAggState{MyString: "created", MySlice: make([]nestedEntity, 0)}, agg.State())
	})

	t.Run(`Given a newly created aggregate
		When Store is called
		And persistFunc returns an error
		Then aggreate's state shouldn't be changed
	`, func(t *testing.T) {
		agg := newTestAgg("id")
		err := agg.Store(func(as AggregateState, ep EventPack, v Version) error {
			return errors.New("error")
		})
		require.Error(t, err)
		require.Equal(t, testAggState{MyString: "created", MySlice: make([]nestedEntity, 0)}, agg.State())
		require.Equal(t, EventPack{Created{}}, agg.events)
		require.Equal(t, Version(0), agg.Version())
	})

	t.Run(`Given an empty aggregate
		When Restore is called
		Then aggreate's state is restored from parmas of Restore
	`, func(t *testing.T) {
		agg := testAgg{}
		id := AggregateId("id")
		state := testAggState{MyString: "created", MySlice: make([]nestedEntity, 0)}
		agg.Restore(id, state, Version(100))
		require.Equal(t, state, agg.State())
		require.Empty(t, agg.events)
		require.Equal(t, Version(100), agg.Version())
	})

	t.Run(`Given a newly created aggregate
		And Error is not nil
		When Restore is called
		Then aggreate's state is restored from parmas of Restore
		And Error is set to nil
	`, func(t *testing.T) {
		id := AggregateId("id")
		agg := newTestAgg("id2")
		agg.err = errors.New("error")
		state := testAggState{MyString: "created", MySlice: make([]nestedEntity, 0)}
		agg.Restore(id, state, Version(100))
		require.Equal(t, id, agg.Id())
		require.Equal(t, state, agg.State())
		require.Empty(t, agg.events)
		require.Equal(t, Version(100), agg.Version())
		require.NoError(t, agg.Error())
	})
}

func BenchmarkAggregate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		agg := newTestAgg("id")
		agg.SingleEventCommand("val")
	}
}
