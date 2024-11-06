package postgres

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"time"

// 	"github.com/gr1nd3rz/go-fast-ddd/core"
// 	"github.com/jackc/pgx/v5"
// 	"github.com/jackc/pgx/v5/pgconn"
// )

// type TableName string

// func NewJsonbPersister[T core.State](table TableName) Persister {
// 	if table == "" {
// 		panic("table name is empty")
// 	}
// 	return &josnbPersister[T]{table}
// }

// type josnbPersister[T core.State] struct {
// 	table TableName
// }

// func (josnbPersister[T]) Match(aggregate any) bool {
// 	_, ok := aggregate.(*core.Aggregate[T])
// 	return ok
// }

// func (p josnbPersister[T]) Restore(ctx context.Context, tx pgx.Tx, id core.Id) (core.State, core.Version, error) {
// 	queryStr := fmt.Sprintf("SELECT version, data FROM %s WHERE id = $1 FOR UPDATE", p.table)
// 	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
// 	defer cancel()
// 	params := []any{id}
// 	row := tx.QueryRow(ctx, queryStr, params...)
// 	var version core.Version
// 	// Add object pooling
// 	var data json.RawMessage
// 	err := row.Scan(&version, &data)
// 	var state T
// 	if err == pgx.ErrNoRows {
// 		return nil, 0, fmt.Errorf("%w [%T] id:[%s]", core.ErrAggregateNotFound, state, id)
// 	}
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	err = json.Unmarshal(data, &state)
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	return state, version, nil
// }

// func (p josnbPersister[T]) Store(ctx context.Context, tx pgx.Tx, id core.Id, state core.State, events core.EventPack, version core.Version) error {
// 	nextVersion := version.Next()
// 	data, err := json.Marshal(state)
// 	if err != nil {
// 		return err
// 	}
// 	if _, err = core.EventOfType[core.Tomstone](events); err == nil {
// 		query := fmt.Sprintf(`DELETE FROM %[1]s WHERE id = $1;`, p.table)
// 		_, err = tx.Exec(ctx, query, id)
// 		return err
// 	}
// 	if version == 0 {
// 		query := fmt.Sprintf(`INSERT INTO %[1]s (id, version, data) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING;`, p.table)
// 		var result pgconn.CommandTag
// 		result, err = tx.Exec(ctx, query, id, nextVersion, data)
// 		if err != nil {
// 			return err
// 		}
// 		rowCount := result.RowsAffected()
// 		if rowCount == 0 {
// 			return fmt.Errorf("ID:[%s] %w", id, core.ErrAggregateExists)
// 		}
// 		return nil
// 	}

// 	query := fmt.Sprintf(`UPDATE %[1]s SET version = $1, data = $2 WHERE version = $3 and id = $4;`, p.table)
// 	result, err := tx.Exec(ctx, query, nextVersion, data, version, id)
// 	if err != nil {
// 		return err
// 	}
// 	rowCount := result.RowsAffected()
// 	if rowCount == 0 {
// 		return fmt.Errorf("ID:[%s] %w", id, core.ErrConcurrentModification)
// 	}
// 	return nil
// }
