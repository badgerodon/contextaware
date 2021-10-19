package contextaware

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"
)

type (
	withSetDeadline interface {
		SetDeadline(time.Time) error
	}
	withSetReadDeadline interface {
		SetReadDeadline(time.Time) error
	}
	withSetWriteDeadline interface {
		SetWriteDeadline(time.Time) error
	}
)

func supportsSetDeadline(obj interface{}) (withSetDeadline, bool) {
	wsd, ok := obj.(withSetDeadline)
	if !ok {
		return nil, false
	}

	err := wsd.SetDeadline(time.Time{})
	return wsd, !errors.Is(err, os.ErrNoDeadline)
}

func supportsSetReadDeadline(obj interface{}) (withSetReadDeadline, bool) {
	wsrd, ok := obj.(withSetReadDeadline)
	if !ok {
		return nil, false
	}

	err := wsrd.SetReadDeadline(time.Time{})
	return wsrd, !errors.Is(err, os.ErrNoDeadline)
}

func supportsSetWriteDeadline(obj interface{}) (withSetWriteDeadline, bool) {
	wswd, ok := obj.(withSetWriteDeadline)
	if !ok {
		return nil, false
	}

	err := wswd.SetWriteDeadline(time.Time{})
	return wswd, !errors.Is(err, os.ErrNoDeadline)
}

// A Reader is an io.Reader that also supports cancellation via a context.Context.
type Reader interface {
	ReadContext(ctx context.Context, p []byte) (n int, err error)
}

type simpleReader struct {
	io.Reader
}

func (r simpleReader) ReadContext(ctx context.Context, p []byte) (n int, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	return r.Read(p)
}

type readerWithSetDeadline struct {
	r           io.Reader
	setDeadline func(time.Time) error
}

func (r readerWithSetDeadline) ReadContext(ctx context.Context, p []byte) (n int, err error) {
	return withCancelViaDeadline(ctx, r.setDeadline, func() (int, error) {
		return r.r.Read(p)
	})
}

// NewReader creates a new contextaware.Reader from an existing io.Reader.
func NewReader(r io.Reader) Reader {
	if cr, ok := r.(Reader); ok {
		return cr
	}
	if obj, ok := supportsSetReadDeadline(r); ok {
		return readerWithSetDeadline{r: r, setDeadline: obj.SetReadDeadline}
	}
	if obj, ok := supportsSetDeadline(r); ok {
		return readerWithSetDeadline{r: r, setDeadline: obj.SetDeadline}
	}
	return simpleReader{r}
}

// A Writer is an io.Writer that also supports cancellation via a context.Context.
type Writer interface {
	WriteContext(ctx context.Context, p []byte) (n int, err error)
}

type simpleWriter struct {
	w io.Writer
}

func (w simpleWriter) WriteContext(ctx context.Context, p []byte) (n int, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	return w.w.Write(p)
}

type writerWithSetDeadline struct {
	w           io.Writer
	setDeadline func(time.Time) error
}

func (w writerWithSetDeadline) WriteContext(ctx context.Context, p []byte) (n int, err error) {
	return withCancelViaDeadline(ctx, w.setDeadline, func() (int, error) {
		return w.w.Write(p)
	})
}

// NewWriter creates a new contextaware.Writer from an existing io.Writer.
func NewWriter(w io.Writer) Writer {
	if cw, ok := w.(Writer); ok {
		return cw
	}
	if obj, ok := supportsSetWriteDeadline(w); ok {
		return writerWithSetDeadline{w: w, setDeadline: obj.SetWriteDeadline}
	}
	if obj, ok := supportsSetDeadline(w); ok {
		return writerWithSetDeadline{w: w, setDeadline: obj.SetDeadline}
	}
	return simpleWriter{w}
}

func withCancelViaDeadline(
	ctx context.Context,
	setDeadline func(time.Time) error,
	operation func() (n int, err error),
) (n int, err error) {
	var cancelOnce sync.Once

	// set the deadline from the context's deadline
	if deadline, ok := ctx.Deadline(); ok {
		err = setDeadline(deadline)
		if err != nil {
			return 0, err
		}
	} else {
		// clear any existing deadline
		err = setDeadline(time.Time{})
		if err != nil {
			return 0, err
		}
	}

	if done := ctx.Done(); done != nil {
		doneCtx, doneCancel := context.WithCancel(ctx)
		defer doneCancel()

		go func() {
			<-doneCtx.Done()
			cancelOnce.Do(func() {
				_ = setDeadline(time.Unix(1, 0))
			})
		}()
	}

	// run the operation
	n, err = operation()

	// Clear any pending cancellation. There are 5 cases:
	//
	// 1. There is no Done() channel
	//    => execute the no-op, there is no background cancel
	// 2. There is a Done() channel, but it hasn't fired yet
	//    => execute the no-op, background cancel won't happen
	// 3. There is a Done() channel, and it has fired, but `cancelOnce` hasn't been called
	//    => execute the no-op, background cancel call will be skipped via sync.Once guarantee
	// 4. There is a Done() channel, and it has fired, and `cancelOnce` is currently being executed
	//    => wait for background cancel call to complete
	// 5. There is a Done() channel, and it has fired, and `cancelOnce` has already executed
	//    => skip call via sync.Once guarantee
	//
	// In all 5 cases we guarantee that the background goroutine is cleaned up and will not call `SetDeadline` after
	// this function returns.
	cancelOnce.Do(func() {})

	return n, err
}
