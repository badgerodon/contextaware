package contextaware

import (
	"context"
	"io"
	"sync"
)

// Pipe creates a synchronous in-memory, context-aware pipe. It is modeled after io.Pipe.
func Pipe() (*PipeReader, *PipeWriter) {
	p := &pipe{
		wrCh: make(chan []byte),
		rdCh: make(chan int),
		done: make(chan struct{}),
	}
	return &PipeReader{p}, &PipeWriter{p}
}

// onceError is an object that will only store an error once.
type onceError struct {
	sync.Mutex // guards following
	err        error
}

func (a *onceError) Store(err error) {
	a.Lock()
	defer a.Unlock()
	if a.err != nil {
		return
	}
	a.err = err
}

func (a *onceError) Load() error {
	a.Lock()
	defer a.Unlock()
	return a.err
}

type pipe struct {
	wrMu sync.Mutex // Serializes Write operations
	wrCh chan []byte
	rdCh chan int

	once sync.Once // Protects closing done
	done chan struct{}
	rerr onceError
	werr onceError
}

func (p *pipe) CloseRead(err error) error {
	if err == nil {
		err = io.ErrClosedPipe
	}
	p.rerr.Store(err)
	p.once.Do(func() { close(p.done) })
	return nil
}

func (p *pipe) CloseWrite(err error) error {
	if err == nil {
		err = io.EOF
	}
	p.werr.Store(err)
	p.once.Do(func() { close(p.done) })
	return nil
}

func (p *pipe) ReadContext(ctx context.Context, b []byte) (n int, err error) {
	select {
	case <-p.done:
		return 0, p.readCloseError()
	default:
	}

	select {
	case bw := <-p.wrCh:
		nr := copy(b, bw)
		p.rdCh <- nr
		return nr, nil
	case <-p.done:
		return 0, p.readCloseError()
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (p *pipe) readCloseError() error {
	rerr := p.rerr.Load()
	if werr := p.werr.Load(); rerr == nil && werr != nil {
		return werr
	}
	return io.ErrClosedPipe
}

func (p *pipe) WriteContext(ctx context.Context, b []byte) (n int, err error) {
	select {
	case <-p.done:
		return 0, p.writeCloseError()
	default:
		p.wrMu.Lock()
		defer p.wrMu.Unlock()
	}

	for once := true; once || len(b) > 0; once = false {
		select {
		case p.wrCh <- b:
			nw := <-p.rdCh
			b = b[nw:]
			n += nw
		case <-p.done:
			return n, p.writeCloseError()
		case <-ctx.Done():
			return n, ctx.Err()
		}
	}
	return n, nil
}

func (p *pipe) writeCloseError() error {
	werr := p.werr.Load()
	if rerr := p.rerr.Load(); werr == nil && rerr != nil {
		return rerr
	}
	return io.ErrClosedPipe
}

type PipeReader struct {
	p *pipe
}

func (pr *PipeReader) Read(data []byte) (n int, err error) {
	return pr.ReadContext(context.Background(), data)
}

func (pr *PipeReader) ReadContext(ctx context.Context, data []byte) (n int, err error) {
	return pr.p.ReadContext(ctx, data)
}

func (pr *PipeReader) Close() error {
	return pr.CloseWithError(nil)
}

func (pr *PipeReader) CloseWithError(err error) error {
	return pr.p.CloseRead(err)
}

type PipeWriter struct {
	p *pipe
}

func (pw *PipeWriter) Write(data []byte) (n int, err error) {
	return pw.WriteContext(context.Background(), data)
}

func (pw *PipeWriter) WriteContext(ctx context.Context, data []byte) (n int, err error) {
	return pw.p.WriteContext(ctx, data)
}

func (pw *PipeWriter) Close() error {
	return pw.CloseWithError(nil)
}

func (pw *PipeWriter) CloseWithError(err error) error {
	return pw.p.CloseWrite(err)
}
