package contextaware

import (
	"context"
	"io"
	"time"
)

// A Reader is an io.Reader that also supports cancellation via a context.Context.
type Reader interface {
	io.Reader
	ReadContext(ctx context.Context, p []byte) (n int, err error)
}

// NewReader creates a new contextaware.Reader from an existing io.Reader.
func NewReader(r io.Reader) Reader {
	return WrapIO(r).(Reader)
}

func wrapReader(r io.Reader) Reader {
	if cr, ok := r.(Reader); ok {
		return cr
	}
	if obj, ok := supportsSetReadDeadline(r); ok {
		return readerViaSetDeadline{r, obj.SetReadDeadline}
	}
	if obj, ok := supportsSetDeadline(r); ok {
		return readerViaSetDeadline{r, obj.SetDeadline}
	}
	return readerViaRead{r}
}

type readerViaRead struct {
	io.Reader
}

func (r readerViaRead) ReadContext(ctx context.Context, p []byte) (n int, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	return r.Read(p)
}

type readerViaSetDeadline struct {
	io.Reader
	setDeadline func(time.Time) error
}

func (r readerViaSetDeadline) ReadContext(ctx context.Context, p []byte) (n int, err error) {
	return withCancelViaDeadline(ctx, r.setDeadline, func() (int, error) {
		return r.Read(p)
	})
}
