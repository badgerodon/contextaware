package contextaware

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapIO(t *testing.T) {
	t.Run("buffer", func(t *testing.T) {
		var buf bytes.Buffer
		ctxbuf := WrapIO(&buf)

		assert.Implements(t, (*io.Reader)(nil), ctxbuf)
		assert.Implements(t, (*io.ReaderFrom)(nil), ctxbuf)
		assert.Implements(t, (*io.Writer)(nil), ctxbuf)
		assert.Implements(t, (*io.WriterTo)(nil), ctxbuf)
	})
	t.Run("reader", func(t *testing.T) {
		r := bytes.NewReader([]byte("EXAMPLE"))
		ctxr := WrapIO(r)

		assert.Implements(t, (*io.Reader)(nil), ctxr)
		assert.Implements(t, (*io.ReaderAt)(nil), ctxr)
		assert.Implements(t, (*io.Seeker)(nil), ctxr)
		assert.Implements(t, (*io.WriterTo)(nil), ctxr)
	})
}
