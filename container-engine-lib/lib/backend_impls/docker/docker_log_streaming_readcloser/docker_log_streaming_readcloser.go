package docker_log_streaming_readcloser

import (
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/sirupsen/logrus"
	"io"
)

// DockerLogStreamingReadCloser is a ReadCloser that demultiplexes a Docker logstream, as Docker log streams
// come in with STDOUT and STDERR multiplexed together and they need to be picked apart by Docker's StdCopy method
type DockerLogStreamingReadCloser struct {
	source io.ReadCloser

	dockerCopyEndedChan chan interface{}

	// The pipe that the Docker demultiplexer will write to
	pipeWriter *io.PipeWriter

	output     *io.PipeReader
}

func NewDockerLogStreamingReadCloser(dockerLogStream io.ReadCloser) *DockerLogStreamingReadCloser {
	pipeReader, pipeWriter := io.Pipe()

	dockerCopyEndedChan := make(chan interface{})
	go func() {
		if _, err := stdcopy.StdCopy(pipeWriter, pipeWriter, dockerLogStream); err != nil {
			logrus.Errorf("An error occurred copying the Docker-multiplexed stream to the pipe: %v", err)
		}
		pipeWriter.Close()
		close(dockerCopyEndedChan)
	}()
	result := &DockerLogStreamingReadCloser{
		source:              dockerLogStream,
		dockerCopyEndedChan: dockerCopyEndedChan,
		output:              pipeReader,
		pipeWriter:          pipeWriter,
	}
	return result
}

func (streamer DockerLogStreamingReadCloser) Read(p []byte) (n int, err error) {
	return streamer.output.Read(p)
}

func (streamer DockerLogStreamingReadCloser) Close() error {
	streamer.source.Close()

	// Wait until the Docker thread exits
	<- streamer.dockerCopyEndedChan

	streamer.output.Close()
	return nil
}

