package core

import (
	"context"
)

type Event any
type EventPack []Event
type AggregateId string
type Version uint64

type AggregateState interface {
	Apply(event Event) AggregateState
}

type Repository interface {
	Load(ctx context.Context, id AggregateId, aggregate IAggregate) error
	Save(ctx context.Context, aggregate IAggregate) error
}
