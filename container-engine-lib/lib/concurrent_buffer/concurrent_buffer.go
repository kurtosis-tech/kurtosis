package concurrent_buffer

import (
	"io"
	"sync"
)

// From https://stackoverflow.com/questions/19646717/is-the-go-bytes-buffer-thread-safe
type concurrentWriter struct {
	underlying io.Writer
	mutex      *sync.Mutex
}
func newConcurrentWriter(underlying io.Writer) *concurrentWriter {
	return &concurrentWriter{
		underlying: underlying,
		mutex:      &sync.Mutex{},
	}
}
func (writer *concurrentWriter) Write(p []byte) (n int, err error) {
	writer.mutex.Lock()
	defer writer.mutex.Unlock()
	return writer.underlying.Write(p)
}
