package core

import "errors"

var ErrConcurrentModification = errors.New("aggregate modified concurrently")
var ErrConcurrentCreation = errors.New("aggregate already exists")
var ErrAggregateNotFound = errors.New("aggregate not found")
var ErrAggregateHasError = errors.New("aggregate has error")
