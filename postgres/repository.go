package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gr1nd3rz/go-fast-ddd/core"
)

var ErrPersisterNotFound = errors.New("persister not found")

type Persister interface {
	Match(aggregate core.IAggregate) bool
	Restore(ctx context.Context, tx *sql.Tx, id core.AggregateId) (core.AggregateState, core.Version, error)
	Store(ctx context.Context, tx *sql.Tx, id core.AggregateId, state core.AggregateState, events core.EventPack, version core.Version) error
}

type postgresRepo struct {
	persisters []Persister
	tx         *sql.Tx
}

var _ core.Repository = (*postgresRepo)(nil)

func (r *postgresRepo) findPersister(aggregate core.IAggregate) Persister {
	for _, persister := range r.persisters {
		if persister.Match(aggregate) {
			return persister
		}
	}
	panic(ErrPersisterNotFound)
}

func (r *postgresRepo) Load(ctx context.Context, id core.AggregateId, aggregate core.IAggregate) error {
	persister := r.findPersister(aggregate)
	state, version, err := persister.Restore(ctx, r.tx, id)
	if err != nil {
		return err
	}
	aggregate.Restore(id, state, version)
	return nil
}

func (r *postgresRepo) Save(ctx context.Context, aggregate core.IAggregate) error {
	persister := r.findPersister(aggregate)
	return aggregate.Store(func(state core.AggregateState, events core.EventPack, version core.Version) error {
		if len(events) == 0 {
			return nil
		}
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		// if aggregate.IsRemoved() {
		// 	fmt.Printf("Deleting item with id %s from %s..\n", state.Id, persister.Table())
		// 	query := fmt.Sprintf(`DELETE FROM %[1]s WHERE id = $1;`, persister.Table())
		// 	_, err := r.tx.ExecContext(ctx, query, state.Id)
		// 	return err
		// }
		// if !hasChanges {
		// return nil
		// }
		return persister.Store(ctx, r.tx, aggregate.Id(), state, events, version)
	})
}

type ConcurrentScope struct {
	db         *sql.DB
	persisters []Persister
}

func NewConcurrentScope(db *sql.DB, persisters []Persister) *ConcurrentScope {
	return &ConcurrentScope{db, persisters}
}

func (c ConcurrentScope) RunTransactional(ctx context.Context, runFunc func(repo core.Repository) error) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	repo := &postgresRepo{c.persisters, tx}
	err = runFunc(repo)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
