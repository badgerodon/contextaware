package contextaware

import (
	"context"
	"errors"
	"io"
)

var errInvalidWrite = errors.New("invalid write result")

// Copy copies from src to dst until either EOF is reached
// on src or an error occurs. It returns the number of bytes
// copied and the first error encountered while copying, if any.
//
// A successful Copy returns err == nil, not err == EOF.
// Because Copy is defined to read from src until EOF, it does
// not treat an EOF from Read as an error to be reported.
//
// This is copied from the stdlib and modified to use contextaware
// Readers and Writers.
func Copy(ctx context.Context, dst Writer, src Reader) (int64, error) {
	return copyBuffer(ctx, dst, src, nil)
}

// copyBuffer is the actual implementation of Copy and CopyBuffer.
// if buf is nil, one is allocated.
//
// This is copied from the stdlib and modified to use contextaware
// Readers and Writers.
func copyBuffer(ctx context.Context, dst Writer, src Reader, buf []byte) (written int64, err error) {
	if buf == nil {
		size := 32 * 1024
		buf = make([]byte, size)
	}
	for {
		nr, er := src.ReadContext(ctx, buf)
		if nr > 0 {
			nw, ew := dst.WriteContext(ctx, buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
