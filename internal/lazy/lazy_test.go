package lazy_test

import (
	"testing"

	"github.com/lewisgibson/go-steam/internal/lazy"
	"github.com/stretchr/testify/require"
)

func TestLazy(t *testing.T) {
	t.Parallel()

	l := lazy.New(func() int {
		return 1
	})

	require.Equal(t, 1, l.Get())
}
