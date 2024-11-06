package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestXxx(t *testing.T) {
	pool, err := pgxpool.New(context.TODO(), "host=localhost port=5432 user=test password=test dbname=db sslmode=disable")
	require.NoError(t, err)
	_, err = pool.Exec(context.TODO(), "insert into checkpoints (consumer_id, commit_id) values ('a', 1)")
	require.NoError(t, err)
	r, err := pool.Query(context.TODO(), "select consumer_id from checkpoints")
	require.NoError(t, err)
	for r.Next() {
		var id string
		err = r.Scan(&id)
		require.NoError(t, err)
		println(id)
	}
}
