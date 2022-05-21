package docker_log_streaming_readcloser

import (
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
)

// DockerLogStreamingReadCloser is a ReadCloser that demultiplexes a Docker logstream, as Docker log streams
// come in with STDOUT and STDERR multiplexed together and they need to be picked apart by Docker's StdCopy method
type DockerLogStreamingReadCloser struct {
	source io.ReadCloser

	dockerCopyEndedChan chan interface{}

	pipeReader *io.PipeReader
	pipeWriter *io.PipeWriter
}

func NewDockerLogStreamingReadCloser(dockerLogStream io.ReadCloser) *DockerLogStreamingReadCloser {
	pipeReader, pipeWriter := io.Pipe()

	dockerCopyEndedChan := make(chan interface{})
	go func() {
		if _, err := stdcopy.StdCopy(pipeWriter, pipeWriter, dockerLogStream); err != nil {
			logrus.Errorf("An error occurred copying the Docker-multiplexed stream to the pipe: %v", err)
		}
		close(dockerCopyEndedChan)
	}()
	result := &DockerLogStreamingReadCloser{
		source:              dockerLogStream,
		dockerCopyEndedChan: dockerCopyEndedChan,
		pipeReader:          pipeReader,
		pipeWriter:          pipeWriter,
	}
	return result
}

func (streamer DockerLogStreamingReadCloser) Read(p []byte) (n int, err error) {
	return streamer.pipeReader.Read(p)
}

func (streamer DockerLogStreamingReadCloser) Close() error {
	if err := streamer.source.Close(); err != nil {
		return stacktrace.Propagate(err, "An error occurred closing the underlying source reader")
	}

	// Wait until the Docker thread exits
	<- streamer.dockerCopyEndedChan

	if err := streamer.pipeWriter.Close(); err != nil {
		return stacktrace.Propagate(err, "An error occurred closing the write end of the pipe")
	}

	if err := streamer.pipeReader.Close(); err != nil {
		return stacktrace.Propagate(err, "An error occurred closing the write end of the pipe")
	}
	return nil
}

