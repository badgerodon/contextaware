package contextaware

import (
	"context"
	"errors"
	"fmt"
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

func withCancelViaDeadline(
	ctx context.Context,
	setDeadline func(time.Time) error,
	operation func() (n int, err error),
) (n int, err error) {
	// fail early
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

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

	switch {
	case err == nil:
		return n, nil
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		return n, fmt.Errorf("%w: %v", context.DeadlineExceeded, err)
	case errors.Is(ctx.Err(), context.Canceled):
		return n, fmt.Errorf("%w: %v", context.Canceled, err)
	default:
		return n, err
	}
}
