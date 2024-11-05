package postgres

import (
	"context"
	"errors"

	"github.com/gr1nd3rz/go-fast-ddd/core"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPersisterNotFound = errors.New("persister not found")

type Persister interface {
	Match(aggregate any) bool
	Restore(ctx context.Context, tx pgx.Tx, id core.Id) (core.State, core.Version, error)
	Store(ctx context.Context, tx pgx.Tx, id core.Id, state core.State, events core.EventPack, version core.Version) error
}

type postgresRepo struct {
	persisters []Persister
	tx         pgx.Tx
}

var _ core.Repository = (*postgresRepo)(nil)

func (r *postgresRepo) findPersister(aggregate any) Persister {
	for _, persister := range r.persisters {
		if persister.Match(aggregate) {
			return persister
		}
	}
	panic(ErrPersisterNotFound)
}

func (r *postgresRepo) Load(ctx context.Context, id core.Id, aggregate core.Restorer) error {
	persister := r.findPersister(aggregate)
	state, version, err := persister.Restore(ctx, r.tx, id)
	if err != nil {
		return err
	}
	aggregate.Restore(id, state, version)
	return nil
}

func (r *postgresRepo) Save(ctx context.Context, aggregate core.Storer) error {
	persister := r.findPersister(aggregate)
	return aggregate.Store(func(id core.Id, state core.State, events core.EventPack, version core.Version) error {
		return persister.Store(ctx, r.tx, id, state, events, version)
	})
}

type ConcurrentScope struct {
	pool       *pgxpool.Pool
	persisters []Persister
}

func NewConcurrentScope(pool *pgxpool.Pool, persisters []Persister) *ConcurrentScope {
	return &ConcurrentScope{pool, persisters}
}

func (c ConcurrentScope) RunTransactional(ctx context.Context, runFunc func(repo core.Repository) error) error {
	return pgx.BeginTxFunc(ctx, c.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		repo := &postgresRepo{c.persisters, tx}
		return runFunc(repo)
	})
}
