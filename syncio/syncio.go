// Package syncio provides syncronised access to io operations
package syncio

import (
	"io"
	"sync"
)

// Writer wraps an io.Writer allowing only syncronous writes to it
type Writer struct {
	mx sync.Mutex
	w  io.Writer
}

// NewWriter returns a new Writer based on the passed io.Writer
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// Write sends the passed []byte to the underlying io.Writer once it has acquired an implicitly associated exclusive lock
func (w *Writer) Write(p []byte) (n int, err error) {
	w.mx.Lock()
	defer w.mx.Unlock()

	return w.w.Write(p)
}
