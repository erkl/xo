package xo

import (
	"io"
)

const minWriteBufSize = 16

// SizedWriter implements the Writer interface using a fixed
// internal buffer of configurable size.
type SizedWriter struct {
	dst io.Writer
	err error
	buf []byte
	n   int
}

var _ Writer = new(SizedWriter)

func NewSizedWriter(w io.Writer, size int) *SizedWriter {
	if size < minWriteBufSize {
		size = minWriteBufSize
	}

	return &SizedWriter{
		dst: w,
		buf: make([]byte, size),
	}
}

func (w *SizedWriter) Write(buf []byte) (int, error) {
	var n int

	// Keep writing to the buffer if there's already some
	// data sitting in it.
	if w.n > 0 {
		n = copy(w.buf[w.n:], buf)
		w.n += n

		// Flush the buffer if we've filled it.
		if w.n == len(w.buf) {
			if err := w.Flush(); err != nil {
				return n, err
			}
		} else {
			return n, nil
		}
	}

	// At this point, the internal buffer is always empty. Bypass it
	// entirely if it's not large enough to hold what data is left.
	if len(buf)-n >= len(w.buf) {
		m, err := w.write(buf[n:])
		return n + m, err
	} else {
		m := copy(w.buf[w.n:], buf[n:])
		w.n += m
		return n + m, nil
	}
}

func (w *SizedWriter) Reserve(n int) ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}

	// If necessary, make room by flushing the buffer.
	if n > len(w.buf)-w.n && w.n > 0 {
		if err := w.Flush(); err != nil {
			return nil, err
		}
	}

	if n > len(w.buf) {
		return nil, ErrCapacity
	} else {
		return w.buf[w.n:], nil
	}
}

func (w *SizedWriter) Commit(n int) error {
	switch {
	case w.err != nil:
		return w.err
	case n <= 0:
		return nil
	case n > len(w.buf)-w.n:
		return errInvalidCommit
	default:
		w.n += n
		return nil
	}
}

func (w *SizedWriter) Flush() error {
	if w.n == 0 {
		return nil
	}

	n, err := w.write(w.buf[:w.n])
	if n > 0 {
		// Recover from short writes.
		if n < w.n {
			w.n -= copy(w.buf[0:], w.buf[n:w.n])
		} else {
			w.n = 0
		}
	}

	return err
}

func (w *SizedWriter) write(buf []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}

	n, err := w.dst.Write(buf)
	if n < len(buf) && err == nil {
		err = io.ErrShortWrite
	} else {
		w.err = err
	}

	return n, err
}
