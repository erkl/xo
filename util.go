package xo

import (
	"bytes"
	"io"
)

// LimitedReader limits the number of bytes which may be read from the
// underlying Reader R to N bytes. With each call to Read or Seek, N is
// updated.
type LimitedReader struct {
	R Reader
	N int64
}

func (lr *LimitedReader) Read(buf []byte) (int, error) {
	if lr.N <= 0 {
		return 0, io.EOF
	}

	if int64(len(buf)) > lr.N {
		buf = buf[:lr.N]
	}

	n, err := lr.R.Read(buf)
	lr.N -= int64(n)
	return n, err
}

func (lr *LimitedReader) Peek(n int) ([]byte, error) {
	var buf []byte
	var err error

	if int64(n) <= lr.N {
		buf, err = lr.R.Peek(n)
	} else {
		// The caller has requested more data than we can provide, so the
		// best case scenario is us failing with io.EOF. However, we should
		// give the underlying data source a chance to fail first.
		buf, err = lr.R.Peek(int(lr.N))
		if err == nil {
			err = io.EOF
		}
	}

	if int64(len(buf)) > lr.N {
		buf = buf[:lr.N]
	}

	return buf, err
}

func (lr *LimitedReader) Seek(n int) error {
	if int64(n) > lr.N {
		return ErrInvalidSeek
	}

	if err := lr.R.Seek(n); err != nil {
		return err
	}

	lr.N -= int64(n)
	return nil
}

// PeekTo peeks further and further into r until it finds the byte c (at an
// index greater than or equal to offset), or r.Peek returns an error.
func PeekTo(r Reader, c byte, offset int) ([]byte, error) {
	for {
		buf, err := r.Peek(offset + 1)
		if err != nil {
			if err == io.EOF && len(buf) > 0 {
				err = io.ErrUnexpectedEOF
			}
			return nil, err
		}

		if i := bytes.IndexByte(buf[offset:], c); i >= 0 {
			return buf[:offset+i+1], nil
		}

		offset = len(buf)
	}
}

// WriteString writes a string to w.
func WriteString(w Writer, s string) (int, error) {
	var n int

	// If s is larger than w's internal buffer, the write has to be
	// split into multiple chunks.
	for len(s) > 0 {
		buf, err := w.Reserve(len(s))
		if len(buf) == 0 {
			if err == nil {
				err = ErrShortReserve
			}
			return n, err
		}

		nc := copy(buf, s)

		if err := w.Commit(nc); err != nil {
			return n, err
		}

		s = s[nc:]
		n += nc
	}

	return n, nil
}
