package core

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
)

type EventRaiser interface {
	Raise(event Event)
	RaisePack(pack EventPack)
	RaiseNotEqual(first any, second any, event Event)
	RaiseTrue(predicate bool, event Event)
}

type applier struct {
	apply func(Event)
}

func (a applier) Raise(event Event) {
	fmt.Printf("EVENT: [%#v]\n", event)
	a.apply(event)
}

func (a applier) RaisePack(pack EventPack) {
	for _, e := range pack {
		a.Raise(e)
	}
}

func (a applier) RaiseNotEqual(first any, second any, event Event) {
	if !cmp.Equal(first, second) {
		a.Raise(event)
	}
}

func (a applier) RaiseTrue(predicate bool, event Event) {
	if predicate {
		a.Raise(event)
	}
}
