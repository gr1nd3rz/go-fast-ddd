package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_core(t *testing.T) {
	scope := NewConcurrentScope(func() Repository { return NewDummyRepository() })
	scope.Run(func(repo Repository) error {
		agg := new(TestAgg)
		repo.Load(context.TODO(), "test", agg)
		require.NotEmpty(t, agg.state)
		return nil
	})
}
