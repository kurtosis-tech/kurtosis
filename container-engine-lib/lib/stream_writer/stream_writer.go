package stream_writer

import (
	"io"
	"sync"
)

// Enables extracting output that is being written to an io.Writer via a channel
type StreamWriter struct {
	underlying    io.Writer
	mutex         *sync.Mutex
	outputChannel chan string
}

func NewStreamWriter(underlying io.Writer, outputChannel chan string) *StreamWriter {
	return &StreamWriter{
		underlying:    underlying,
		mutex:         &sync.Mutex{},
		outputChannel: outputChannel,
	}
}

func (writer *StreamWriter) Write(p []byte) (n int, err error) {
	writer.mutex.Lock()
	defer writer.mutex.Unlock()

	writer.outputChannel <- string(p)
	return writer.underlying.Write(p)
}
