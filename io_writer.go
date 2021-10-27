package contextaware

import (
	"context"
	"io"
	"time"
)

// A Writer is an io.Writer that also supports cancellation via a context.Context.
type Writer interface {
	io.Writer
	WriteContext(ctx context.Context, p []byte) (n int, err error)
}

// NewWriter creates a new contextaware.Writer from an existing io.Writer.
func NewWriter(w io.Writer) Writer {
	return WrapIO(w).(Writer)
}

// wrapWriter creates a new contextaware.Writer from an existing io.Writer.
func wrapWriter(w io.Writer) Writer {
	if cw, ok := w.(Writer); ok {
		return cw
	}
	if obj, ok := supportsSetWriteDeadline(w); ok {
		return writerViaSetDeadline{w, obj.SetWriteDeadline}
	}
	if obj, ok := supportsSetDeadline(w); ok {
		return writerViaSetDeadline{w, obj.SetDeadline}
	}
	return writerViaWrite{w}
}

type writerViaWrite struct {
	io.Writer
}

func (w writerViaWrite) WriteContext(ctx context.Context, p []byte) (n int, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	return w.Write(p)
}

type writerViaSetDeadline struct {
	io.Writer
	setDeadline func(time.Time) error
}

func (w writerViaSetDeadline) WriteContext(ctx context.Context, p []byte) (n int, err error) {
	return withCancelViaDeadline(ctx, w.setDeadline, func() (int, error) {
		return w.Write(p)
	})
}
