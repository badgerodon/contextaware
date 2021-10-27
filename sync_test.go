package contextaware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMutex(t *testing.T) {
	var m Mutex
	m.Lock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.ErrorIs(t, context.Canceled, m.LockContext(ctx))

	m.Unlock()

	assert.NoError(t, m.LockContext(context.Background()))
}
