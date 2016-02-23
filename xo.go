// Package xo implements buffered I/O for Go.
package xo

import (
	"errors"
	"io"
)

var (
	// ErrBufferTooSmall may be used by Reader or Writer implementations to
	// indicate that their internal buffers are too small to fulfill a Peek or
	// Reserve request.
	ErrBufferTooSmall = errors.New("xo: buffer too small")

	// ErrShortPeek and ErrShortReserve describe the case that Reader.Peek
	// or Writer.Reserve returned a smaller byte slice than requested, without
	// also providing an error explaining why.
	ErrShortPeek    = errors.New("xo: short peek")
	ErrShortReserve = errors.New("xo: short reserve")

	// ErrInvalidConsumeSize and ErrInvalidCommitSize may be used by Reader or
	// Writer implementations to indicate that the size argument in a call to
	// Consume or Commit is invalid.
	ErrInvalidConsumeSize = errors.New("xo: invalid consume size")
	ErrInvalidCommitSize  = errors.New("xo: invalid commit size")
)

type Reader interface {
	io.Reader

	// Peek returns at least n bytes of unread bytes from the Reader's internal
	// buffer, without consuming them, reading more data into the internal
	// buffer first if necessary. The byte slice is only valid until the next
	// read operation.
	//
	// If Peek returns fewer than n bytes, it must also return an error
	// explaining why.
	Peek(n int) ([]byte, error)

	// Consume discards n bytes from the Reader's internal buffer.
	Consume(n int) error
}

type Writer interface {
	io.Writer

	// Reserve returns at least n bytes of scratch space from the Writer's
	// internal buffer, flushing data to make room if necessary. The scratch
	// space is only valid until the next write operation.
	//
	// If Reserve returns fewer than n bytes, it must also return an error
	// explaining why.
	Reserve(n int) ([]byte, error)

	// Commit commits the first n bytes of scratch space to be written.
	Commit(n int) error

	// Flush writes all buffered data.
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
