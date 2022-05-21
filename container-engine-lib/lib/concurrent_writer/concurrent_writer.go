package concurrent_writer

import (
	"io"
	"sync"
)

// From https://stackoverflow.com/questions/19646717/is-the-go-bytes-buffer-thread-safe
type ConcurrentWriter struct {
	underlying io.Writer
	mutex      *sync.Mutex
}
func NewConcurrentWriter(underlying io.Writer) *ConcurrentWriter {
	return &ConcurrentWriter{
		underlying: underlying,
		mutex:      &sync.Mutex{},
	}
}
func (writer *ConcurrentWriter) Write(p []byte) (n int, err error) {
	writer.mutex.Lock()
	defer writer.mutex.Unlock()
	return writer.underlying.Write(p)
}
