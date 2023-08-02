package stream_writer

import (
	"io"
	"sync"
	"time"
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

	// TODO: remove this time.Sleep
	// This sleep is a hack to make the streamed output UI deliver steadily as its
	// should be removed once UX of streaming is improved
	time.Sleep(500 * time.Millisecond)
	return writer.underlying.Write(p)
}
