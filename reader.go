package xo

import (
	"io"
)

const minReadBufSize = 4096

// GrowingReader is a Reader implementation which will resize its
// internal buffer automatically to accomodate larger Peek calls.
type GrowingReader struct {
	src  io.Reader
	buf  []byte
	r, w int
}

var _ Reader = new(GrowingReader)

func NewGrowingReader(r io.Reader) *GrowingReader {
	return &GrowingReader{src: r}
}

func (r *GrowingReader) Read(buf []byte) (int, error) {
	if r.r == r.w {
		if len(buf) >= len(r.buf) {
			return r.read(buf)
		}
		if err := r.fill(1); err != nil {
			return 0, err
		}
	}

	n := copy(buf[0:], r.buf[r.r:r.w])
	r.r += n
	return n, nil
}

func (r *GrowingReader) Peek(n int) ([]byte, error) {
	if err := r.fill(n); err != nil {
		return nil, err
	} else {
		return r.buf[r.r:r.w], nil
	}
}

func (r *GrowingReader) Discard(n int) error {
	if n > r.w-r.r {
		return errInvalidDiscard
	}
	if n > 0 {
		r.r += n
	}
	return nil
}

func (r *GrowingReader) Shrink() {
	// Don't bother trying to shrink a buffer that hasn't
	// been created yet.
	if len(r.buf) == 0 {
		return
	}

	n := idealSize(len(r.buf), r.w-r.r)
	if n != len(r.buf) && n > 0 {
		r.resize(n)
	}
}

func (r *GrowingReader) read(buf []byte) (int, error) {
	// To deal with weird io.Reader implementations, try calling
	// Read operation a number of times before giving up.
	for i := 0; i < 10; i++ {
		n, err := r.src.Read(buf)
		if n > 0 {
			return n, nil
		} else if err != nil {
			return 0, err
		}
	}

	return 0, io.ErrNoProgress
}

func (r *GrowingReader) fill(n int) error {
	// Make sure there's enough free space after the already
	// consumed portion of the buffer (i.e. n <= len(r.buf)-r.r).
	if n > len(r.buf)-r.r {
		if len(r.buf) < n {
			m := idealSize(len(r.buf), n)
			if m <= 0 {
				return ErrCapacity
			} else {
				r.resize(m)
			}
		} else if r.r == r.w {
			r.r, r.w = 0, 0
		} else {
			r.r, r.w = 0, copy(r.buf[0:], r.buf[r.r:r.w])
		}
	}

	// Repeatedly request more data from our source until we have
	// at least n bytes sitting in the buffer.
	for r.w-r.r < n {
		nr, err := r.read(r.buf[r.w:])
		if err != nil {
			return err
		}
		r.w += nr
	}

	return nil
}

func (r *GrowingReader) resize(n int) {
	buf := make([]byte, n)
	r.r = copy(buf[0:], r.buf[r.r:r.w])
	r.w = 0
	r.buf = buf
}

func idealSize(n, min int) int {
	if n < minReadBufSize {
		n = minReadBufSize
	}

	for n < min {
		if x := n << 1; x>>1 == n {
			return -1
		} else {
			n = x
		}
	}

	return n
}
