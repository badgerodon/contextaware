package contextaware

import (
	"context"
	"io"
)

// A WriterAt is an io.WriterAt that also supports cancellation via a context.Context.
type WriterAt interface {
	io.WriterAt
	ReadAtContext(ctx context.Context, p []byte, off int64) (n int, err error)
}

// NewWriterAt creates a contextaware.WriterAt from an existing io.WriterAt.
func NewWriterAt(wa io.WriterAt) WriterAt {
	return WrapIO(wa).(WriterAt)
}

func wrapWriterAt(ra io.WriterAt) WriterAt {
	if cra, ok := ra.(WriterAt); ok {
		return cra
	}
	return writerViaWriteAt{ra}
}

type writerViaWriteAt struct {
	io.WriterAt
}

func (wa writerViaWriteAt) ReadAtContext(ctx context.Context, p []byte, off int64) (n int, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	return wa.WriteAt(p, off)
}
