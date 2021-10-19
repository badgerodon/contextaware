package contextaware

import (
	"context"
	"sync"
)

// A Locker is a context-aware Locker.
type Locker interface {
	LockContext(context.Context) error
	Unlock()
}

// A Mutex implements the contextaware.Locker and io.Locker interfaces. It is *not* reentrant.
type Mutex struct {
	once sync.Once
	ch   chan struct{}
}

// Lock locks the mutex using the background context.
func (mu *Mutex) Lock() {
	_ = mu.LockContext(context.Background())
}

// LockContext locks the mutex. If `ctx.Done()` fires before a lock is acquired an error will be returned. If the lock
// was successfully taken, nil will be returned.
func (mu *Mutex) LockContext(ctx context.Context) error {
	mu.init()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-mu.ch:
		return nil
	}
}

// Unlock unlocks the mutex. Only a locked mutex may be unlocked.
func (mu *Mutex) Unlock() {
	mu.init()

	select {
	case mu.ch <- struct{}{}:
	default:
		panic("contextaware: unlock of unlocked mutex")
	}
}

func (mu *Mutex) init() {
	mu.once.Do(func() {
		mu.ch = make(chan struct{}, 1)
		mu.ch <- struct{}{}
	})
}
