package core

import (
	"context"
	"errors"
)

var ErrConcurrentModification = errors.New("aggregate modified concurrently")
var ErrConcurrentCreation = errors.New("aggregate already exists")
var ErrAggregateNotFound = errors.New("aggregate not found")

type Repository interface {
	Load(ctx context.Context, id Id, aggregate Restorer) error
	Save(ctx context.Context, aggregate Storer) error
}
