package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gr1nd3rz/go-fast-ddd/core"
)

type TableName string

func NewJsonbPersister[T core.AggregateState](table TableName) Persister {
	if table == "" {
		panic("table is empty")
	}
	return &josnbPersister[T]{table}
}

type josnbPersister[T core.AggregateState] struct {
	table TableName
}

func (josnbPersister[T]) Match(aggregate core.IAggregate) bool {
	_, ok := aggregate.(*core.Aggregate[T])
	return ok
}

func (p josnbPersister[T]) Restore(ctx context.Context, tx *sql.Tx, id core.AggregateId) (core.AggregateState, core.Version, error) {
	queryStr := fmt.Sprintf("SELECT version, data FROM %s WHERE id = $1 FOR UPDATE", p.table)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	params := []any{id}
	row := tx.QueryRowContext(ctx, queryStr, params...)
	var version core.Version
	var data json.RawMessage
	err := row.Scan(&version, &data)
	var state T
	if err == sql.ErrNoRows {
		return nil, 0, fmt.Errorf("%w [%T] id:[%s]", core.ErrAggregateNotFound, state, id)
	}
	if err != nil {
		return nil, 0, err
	}
	err = json.Unmarshal(data, &state)
	if err != nil {
		return nil, 0, err
	}
	return state, version, nil
}

func (p josnbPersister[T]) Store(ctx context.Context, tx *sql.Tx, id core.AggregateId, state core.AggregateState, events core.EventPack, version core.Version) error {
	nextVersion := version + 1
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if version == 0 {
		query := fmt.Sprintf(`INSERT INTO %[1]s (id, version, data) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING;`, p.table)
		result, err := tx.ExecContext(ctx, query, id, nextVersion, data)
		if err != nil {
			return err
		}
		rowCount, err := result.RowsAffected()
		if err != nil {
			panic(err)
		}
		if rowCount == 0 {
			return core.ErrConcurrentCreation
		}
	} else {
		query := fmt.Sprintf(`UPDATE %[1]s SET version = $1, data = $2 WHERE version = $3 and id = $4;`, p.table)
		result, err := tx.ExecContext(ctx, query, nextVersion, data, version, id)
		if err != nil {
			return err
		}
		rowCount, err := result.RowsAffected()
		if err != nil {
			panic(err)
		}
		if rowCount == 0 {
			return core.ErrConcurrentModification
		}
	}
	return nil
}
