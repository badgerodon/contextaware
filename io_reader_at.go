package contextaware

import (
	"context"
	"io"
)

// A ReaderAt is an io.ReaderAt that also supports cancellation via a context.Context.
type ReaderAt interface {
	io.ReaderAt
	ReadAtContext(ctx context.Context, p []byte, off int64) (n int, err error)
}

// NewReaderAt creates a contextaware.ReaderAt from an existing io.ReaderAt.
func NewReaderAt(ra io.ReaderAt) ReaderAt {
	return WrapIO(ra).(ReaderAt)
}

func wrapReaderAt(ra io.ReaderAt) ReaderAt {
	if cra, ok := ra.(ReaderAt); ok {
		return cra
	}
	return readerAtViaReadAt{ra}
}

type readerAtViaReadAt struct {
	io.ReaderAt
}

func (ra readerAtViaReadAt) ReadAtContext(ctx context.Context, p []byte, off int64) (n int, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	return ra.ReadAt(p, off)
}
