// Package xo implementes buffered I/O.
package xo

import (
	"errors"
	"io"
)

var (
	// ErrCapacity may be returned by Peek or Reserve when a Reader or
	// Writer's internal buffer is too small to satisfy the request.
	ErrCapacity = errors.New("xo: insufficient buffer capacity")

	// ErrShortPeek and ErrShortReserve signal that a Peek or Reserve call
	// yielded fewer bytes than expected but failed to return an explicit
	// error.
	ErrShortPeek    = errors.New("xo: short peek")
	ErrShortReserve = errors.New("xo: short reserve")

	errInvalidDiscard = errors.New("xo: invalid discard")
	errInvalidCommit  = errors.New("xo: invalid commit")
)

type Reader interface {
	io.Reader

	// Peek returns at least n bytes of unread data from the Reader's internal
	// buffer without consuming them, reading more data into the internal
	// buffer first if necessary. The byte slice is only valid until the next
	// Read or Discard operation.
	//
	// Implementations must return a slice of at least n bytes (but preferably
	// as much as possible), or a non-nil error.
	Peek(n int) ([]byte, error)

	// Discard drops the first n bytes from the reader's internal buffer.
	Discard(n int) error
}

type Writer interface {
	io.Writer

	// Reserve allocates n or more bytes from the Writer's internal buffer
	// to be used as scratch space, flushing existing data first to make room
	// if necessary. The byte slice is only valid until the next Write or
	// Commit call.
	//
	// Implementations must return a slice of at least n bytes (but preferably
	// as much as possible), or a non-nil error.
	Reserve(n int) ([]byte, error)

	// Commit commits the first n bytes of scratch space previously returned
	// by a Reserve call, to the writer's internal buffer.
	Commit(n int) error

	// Flush flushes all buffered data to the destination io.Writer.
	Flush() error
}

// ReadWriter combines the Reader and Writer interfaces.
type ReadWriter interface {
	Reader
	Writer
}

// NewReadWriter combines a Reader and a Writer into a ReadWriter.
func NewReadWriter(r Reader, w Writer) ReadWriter {
	return &rw{r, w}
}

type rw struct {
	Reader
	Writer
}
