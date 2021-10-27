package contextaware

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadWrite(t *testing.T) {
	t.Run("deadline", func(t *testing.T) {
		ctx, clearTimeout := context.WithTimeout(context.Background(), -1)
		defer clearTimeout()

		var buf bytes.Buffer
		p := make([]byte, 4)
		n, err := NewReader(&buf).ReadContext(ctx, p)
		assert.Equal(t, 0, n)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		n, err = NewWriter(&buf).WriteContext(ctx, p)
		assert.Equal(t, 0, n)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
	t.Run("cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		var buf bytes.Buffer
		p := make([]byte, 4)
		n, err := NewReader(&buf).ReadContext(ctx, p)
		assert.Equal(t, 0, n)
		assert.ErrorIs(t, err, context.Canceled)
		n, err = NewWriter(&buf).WriteContext(ctx, p)
		assert.Equal(t, 0, n)
		assert.ErrorIs(t, err, context.Canceled)
	})
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		var buf bytes.Buffer
		n, err := NewWriter(&buf).WriteContext(ctx, []byte{1, 2, 3, 4})
		assert.Equal(t, 4, n)
		assert.NoError(t, err)
		p := make([]byte, 4)
		n, err = NewReader(&buf).ReadContext(ctx, p)
		assert.Equal(t, 4, n)
		assert.NoError(t, err)
		assert.Equal(t, []byte{1, 2, 3, 4}, p)
	})
}

func TestDeadlineReadWrite(t *testing.T) {
	li, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer li.Close()

	c1, err := net.Dial("tcp", li.Addr().String())
	require.NoError(t, err)

	c2, err := li.Accept()
	require.NoError(t, err)

	cr := NewReader(c1)
	cw := NewWriter(c2)

	t.Run("deadline", func(t *testing.T) {
		ctx, clearTimeout := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer clearTimeout()

		p := make([]byte, 4)
		n, err := cr.ReadContext(ctx, p)
		assert.Equal(t, 0, n)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		n, err = cw.WriteContext(ctx, p)
		assert.Equal(t, 0, n)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
	t.Run("cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(time.Millisecond*100, func() {
			cancel()
		})

		p := make([]byte, 4)
		n, err := cr.ReadContext(ctx, p)
		assert.Equal(t, 0, n)
		assert.ErrorIs(t, err, context.Canceled)
		n, err = cw.WriteContext(ctx, p)
		assert.Equal(t, 0, n)
		assert.ErrorIs(t, err, context.Canceled)
	})
}
